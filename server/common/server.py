import socket
import logging
import signal
import threading
import protocol
from common import utils


class Server:


    def __init__(self, port, listen_backlog, agency_amount):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        self._server_socket.settimeout(1.0)
        self._threads = []
        self._clients_sockets = {}
        self._clients_lock = threading.Lock()
        self._storage_lock = threading.Lock()
        self._barrier = threading.Barrier(agency_amount, action=self.__on_draw_ready)
        self._winners_by_agency = {}
        self._winners_lock = threading.Lock()


        self._stop_event = threading.Event()
        signal.signal(signal.SIGTERM, self.sigterm_handler)
        signal.signal(signal.SIGINT, self.sigterm_handler)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while not self._stop_event.is_set():
            try:
                client_sock = self.__accept_new_connection()
            except socket.timeout:
                continue
            except OSError:
                break
            
            client_thread = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
            client_thread.start()
            key = client_sock.fileno()
            with self._clients_lock:
                self._clients_sockets[key] = client_sock
            self._threads.append(client_thread)
        
    def sigterm_handler(self, signum, frame):
        logging.info("action: shutdown | result: in_progress")
        self._stop_event.set()

        try:
            self._barrier.abort()
        except Exception:
            pass

        self._server_socket.close()
        
        with self._clients_lock:
            for s in self._clients_sockets.values():
                s.close()

        for t in self._threads:
            t.join()
        logging.info("action: shutdown | result: success")
    
    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            addr = client_sock.getpeername()
            total_bets = self.__process_bet_messages(client_sock)
            logging.info(f"action: process_bets | result: success | ip: {addr[0]} | cantidad: {total_bets}")

            self._barrier.wait(timeout=30)

            self.__respond_to_winner_request(client_sock)
            logging.info(f"action: send_winner_response | result: success | ip: {addr[0]}")

        except ValueError as e:
            logging.error(f"action: apuesta_recibida | result: fail | ip: {addr[0]} | error: {e}")
            try:
                protocol.send_message(client_sock, protocol.TYPE_ACK, 0, False, b"1")
            except Exception:
                pass
            self._barrier.abort()
        except threading.BrokenBarrierError:
            logging.error(f"action: sorteo | result: fail | ip: {addr[0]} | error: Barrier broken")
        except ConnectionError:
            logging.info(f"action: connection_closed | result: in_progress | ip: {addr[0]}")
            self._barrier.abort()
        except OSError as e:
            logging.error(f"action: connection_error | result: fail | ip: {addr[0]} | error: {e}")
            self._barrier.abort()
        finally:
            key = client_sock.fileno()
            client_sock.close()
            with self._clients_lock:
                self._clients_sockets.pop(key, None)
            logging.info(f"action: connection_closed | result: success | ip: {addr[0]}")


    def __process_bet_messages(self, client_sock):
        addr = client_sock.getpeername()
        total_bets = 0
        eof = False
        while not eof:
            msg_type, agency_id, eof_flag, payload = protocol.read_message(client_sock)
            if msg_type != protocol.TYPE_BET:
                raise ValueError(f"Expected bet type {protocol.TYPE_BET}, got {msg_type}")
                
            bets = utils.deserialize_bets(agency_id, payload)
            len_bets = len(bets)
                
            logging.info(f'action: store_bets | result: in_progress | ip: {addr[0]}')
            with self._storage_lock:
                utils.store_bets(bets)
            logging.info(f'action: store_bets | result: success | ip: {addr[0]}')
                
            protocol.send_message(client_sock, protocol.TYPE_ACK, 0, False, b"0")
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len_bets} | ip: {addr[0]}')
            eof = eof_flag
            total_bets += len_bets

        return total_bets
    
    def __on_draw_ready(self):
        logging.info("action: sorteo | result: success")
        winners_by_agency = {}
        with self._storage_lock:
            for bet in utils.load_bets():
                if utils.has_won(bet):
                    winners_by_agency.setdefault(bet.agency, []).append(bet)
        
        with self._winners_lock:
            self._winners_by_agency = winners_by_agency

    def __respond_to_winner_request(self, client_sock):
        msg_type, agency_id, _, _ = protocol.read_message(client_sock)
        if msg_type != protocol.TYPE_WINNER_QUERY:
            raise ValueError(f"Expected winner query type {protocol.TYPE_WINNER_QUERY}, got {msg_type}")

        with self._winners_lock:
            winners = self._winners_by_agency.get(agency_id, [])
        if len(winners) == 0:
            payload = b""
        else:
            payload = "\n".join([f"{bet.document}" for bet in winners]).encode('utf-8')

        protocol.send_message(client_sock, protocol.TYPE_WINNER_RESPONSE, 0, True, payload)
