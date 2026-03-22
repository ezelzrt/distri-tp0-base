package protocol

import (
    "encoding/binary"
    "errors"
    "io"
    "net"
)

type MessageType byte

const (
    BetType MessageType = 0
    AckType     MessageType = 1
)

type Packet struct {
    Type    MessageType
    Payload []byte
}

const MaxPacketSize = 8 * 1024 // 8kb
const HeaderSize = 1 + 4 // type + payload length
const MaxPayloadSize = MaxPacketSize - HeaderSize

func encodeMsg(msgType MessageType, payload []byte) []byte {
    payloadLen := uint32(len(payload))
    buf := make([]byte, HeaderSize+len(payload))

    buf[0] = byte(msgType)
    binary.BigEndian.PutUint32(buf[1:5], payloadLen)
    copy(buf[5:], payload)

    return buf
}

func decodeHeader(header []byte) (MessageType, uint32, error) {
    if len(header) != HeaderSize {
        return 0, 0, errors.New("header size mismatch")
    }
    msgType := MessageType(header[0])
    length := binary.BigEndian.Uint32(header[1:5])
    return msgType, length, nil
}

// Se asegura de escribir todo el buffer en la conexión
func writeAll(w io.Writer, data []byte) error {
    total := 0
    for total < len(data) {
        n, err := w.Write(data[total:])
        if err != nil {
            return err
        }
        total += n
    }
    return nil
}

// Lee el mensaje completo, header + payload
func ReadMsg(conn net.Conn) (MessageType, []byte, error) {
    header := make([]byte, HeaderSize)
    if _, err := io.ReadFull(conn, header); err != nil {
        return 0, nil, err
    }

    msgType, length, err := decodeHeader(header)
    if err != nil {
        return 0, nil, err
    }

    payload := make([]byte, length)
    if _, err := io.ReadFull(conn, payload); err != nil {
        return 0, nil, err
    }

    return msgType, payload, nil
}

// Envia el paquete con el protocolo definido
func SendMsg(conn net.Conn, msgType MessageType, payload []byte) error {
    packet := encodeMsg(msgType, payload)
    return writeAll(conn, packet)
}