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
    WinnerQueryType MessageType = 2
    WinnerResponseType MessageType = 3
)

type Packet struct {
    Type    MessageType
    Payload []byte
}

const MaxPacketSize = 8 * 1024 // 8kb
const HeaderSize = 1 + 2 + 1 + 4 // type + agency ID + EOF flag + payload length
const MaxPayloadSize = MaxPacketSize - HeaderSize

func encodeMsg(msgType MessageType, agencyID uint16, eofFlag byte, payload []byte) []byte {
    payloadLen := uint32(len(payload))
    buf := make([]byte, HeaderSize+len(payload))

    buf[0] = byte(msgType)
    binary.BigEndian.PutUint16(buf[1:3], agencyID)
    buf[3] = eofFlag
    binary.BigEndian.PutUint32(buf[4:8], payloadLen)
    copy(buf[8:], payload)

    return buf
}

func decodeHeader(header []byte) (MessageType, uint16, bool, uint32, error) {
    if len(header) != HeaderSize {
        return 0, 0, false, 0, errors.New("header size mismatch")
    }
    msgType := MessageType(header[0])
    agencyID := binary.BigEndian.Uint16(header[1:3])
    eofFlag := header[3]
    length := binary.BigEndian.Uint32(header[4:8])
    return msgType, agencyID, uint32(eofFlag) == 1, length, nil
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
func ReadMsg(conn net.Conn) (MessageType, uint16, bool, []byte, error) {
    header := make([]byte, HeaderSize)
    if _, err := io.ReadFull(conn, header); err != nil {
        return 0, 0, false, nil, err
    }

    msgType, agencyID, eofFlag, length, err := decodeHeader(header)
    if err != nil {
        return 0, 0, false, nil, err
    }

    payload := make([]byte, length)
    if _, err := io.ReadFull(conn, payload); err != nil {
        return 0, 0, false, nil, err
    }

    return msgType, agencyID, eofFlag, payload, nil
}

// Envia el paquete con el protocolo definido
func SendMsg(conn net.Conn, msgType MessageType, agencyID uint16, eofFlag bool, payload []byte) error {
    eofFlagByte := byte(0)
    if eofFlag {
        eofFlagByte = 1
    }
    packet := encodeMsg(msgType, agencyID, eofFlagByte, payload)
    return writeAll(conn, packet)
}