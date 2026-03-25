import struct
import socket

# Types
TYPE_BET = 0
TYPE_ACK = 1
TYPE_WINNER_QUERY = 2
TYPE_WINNER_RESPONSE = 3

HEADER_FMT = "!BHBI"  # type (1 byte), agency ID (2 bytes), EOF flag (1 byte), payload length (4 bytes)
HEADER_SIZE = struct.calcsize(HEADER_FMT)
MAX_PACKET_SIZE = 8 * 1024
MAX_PAYLOAD_SIZE = MAX_PACKET_SIZE - HEADER_SIZE

def recv_all(sock: socket.socket, n: int) -> bytes:
    data = bytearray()
    while len(data) < n:
        chunk = sock.recv(n - len(data))
        if not chunk:
            raise ConnectionError("Connection closed during recv")
        data.extend(chunk)
    return bytes(data)

def send_all(sock: socket.socket, data: bytes) -> None:
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise ConnectionError("Connection closed during send")
        total_sent += sent

def pack_message(msg_type: int, agency_id: int, eof_flag: int, payload: bytes) -> bytes:
    length = len(payload)
    header = struct.pack(HEADER_FMT, msg_type, agency_id, eof_flag, length)
    return header + payload

def unpack_header(header: bytes):
    if len(header) != HEADER_SIZE:
        raise ValueError("Invalid header length")
    msg_type, agency_id, eof_flag, length = struct.unpack(HEADER_FMT, header)
    return msg_type, agency_id, eof_flag, length

def send_message(sock: socket.socket, msg_type: int, agency_id: int, eof_flag: bool, payload: bytes):
    packet = pack_message(msg_type, agency_id, int(eof_flag), payload)
    send_all(sock, packet)

def read_message(sock: socket.socket):
    header = recv_all(sock, HEADER_SIZE)
    msg_type, agency_id, eof_flag, length = unpack_header(header)
    payload = recv_all(sock, length) if length > 0 else b""
    return msg_type, agency_id, eof_flag == 1, payload
