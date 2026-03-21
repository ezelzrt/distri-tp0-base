import struct
import socket

# Types
TYPE_BET = 0
TYPE_ACK = 1

HEADER_FMT = "!BI"  # Big-endian: unsigned byte + unsigned int
HEADER_SIZE = struct.calcsize(HEADER_FMT)  # 5 bytes

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

def pack_message(msg_type: int, payload: bytes) -> bytes:
    length = len(payload)
    header = struct.pack(HEADER_FMT, msg_type, length)
    return header + payload

def unpack_header(header: bytes):
    if len(header) != HEADER_SIZE:
        raise ValueError("Invalid header length")
    msg_type, length = struct.unpack(HEADER_FMT, header)
    return msg_type, length

def send_message(sock: socket.socket, msg_type: int, payload: bytes):
    packet = pack_message(msg_type, payload)
    send_all(sock, packet)

def read_message(sock: socket.socket):
    header = recv_all(sock, HEADER_SIZE)
    msg_type, length = unpack_header(header)
    payload = recv_all(sock, length) if length > 0 else b""
    return msg_type, payload
