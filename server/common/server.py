import socket
import logging
import signal
import threading
import protocol
from common import utils


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        self._server_socket.settimeout(1.0)
        self._threads = []
        self._clients_sockets = {}
        self._clients_lock = threading.Lock()
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
        self._server_socket.close()
        
        with self._clients_lock:
            for s in self._clients_sockets.values():
                s.close()

        for t in self._threads:
            t.join()
        logging.info("action: shutdown | result: success")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:

            msg_type, payload = protocol.read_message(client_sock)
            if msg_type != protocol.TYPE_BET:
                raise ValueError(f"Expected bet type {protocol.TYPE_BET}, got {msg_type}")
            addr = client_sock.getpeername()
            
            bets = utils.deserialize_bets(payload)
            logging.info(f'action: store_bets | result: in_progress | ip: {addr[0]}')
            utils.store_bets(bets)
            logging.info(f'action: store_bets | result: success | ip: {addr[0]}')
            protocol.send_message(client_sock, protocol.TYPE_ACK, b"0")
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

        except ValueError as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: {len(bets)}")
            try:
                protocol.send_message(client_sock, protocol.TYPE_ACK, b"1")
            except Exception:
                pass

        except OSError as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            try:
                protocol.send_message(client_sock, protocol.TYPE_ACK, b"1")
            except Exception:
                pass

        finally:
            key = client_sock.fileno()
            client_sock.close()
            with self._clients_lock:
                self._clients_sockets.pop(key, None)

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
