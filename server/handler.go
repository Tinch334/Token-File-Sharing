package server


import (
    "fmt"
	"strings"
    "errors"
    "time"
    "slices"
    "context"
    "net"
    "io"
    "encoding/binary"

	"github.com/rs/zerolog/log"
)


// A type for packet type handlers, they take the packets data and return the response that should be sent to the connected host.
// The error string should contain a description of the error.
type TypeHandler func([][]byte) ([]byte, error)

var handlerMap = map[MessageCode]TypeHandler{
    ERR:      dummy,
    USR_CONN: dummy,
    TOK_CONN: dummy,
    ECHO_S:   echoHandler,
    ECHO_R:   dummy
    FLE_LST:  dummy,
}


// Handles incoming connections to the server.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) (int, error) {
    // Set read deadline in case connected host disconnects.
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

    var err error

    // Read packet type.
    header := make([]byte, 1)
    _, err = io.ReadFull(conn, header)

    if err != nil {
        // Invalid packet size.
        if err == io.ErrUnexpectedEOF {
            // Send error message.
            
        }
        conn.Close()
        return 0, err
    }

    // Stores all data blocks.
    var data [][]byte

    for {
        // Read packet data size.
        packetSizeBytes := make([]byte, 4)
        _, err = io.ReadFull(conn, packetSizeBytes)

        if err != nil {
            // Invalid packet size.
            if err == io.ErrUnexpectedEOF {
                // Send error message.
                
            }
            conn.Close()
            return 0, err
        }

        // Read packet data.
        packetSize := binary.BigEndian.Uint32(packetSizeBytes)
        packet := make([]byte, packetSize)
        _, err = io.ReadFull(conn, packet)

        if err != nil {
            // Invalid packet size.
            if err == io.ErrUnexpectedEOF {
                // Send error message.
                
            }
            conn.Close()
            return 0, err
        }

        // Store read packet data.
        data = append(data, packet)

        // Read additional block indicator.
        aBlock := make([]byte, 1)
        _, err = io.ReadFull(conn, aBlock)

        if err != nil {
            // Invalid packet size.
            if err == io.ErrUnexpectedEOF {
                // Send error message.
                
            }
            conn.Close()
            return 0, err
        }

        // Check if there's more data to be read.
        if aBlock == ADDITIONAL_BLOCK_NO {
            break
        }
    }



    return 0, nil
}

// printf '\x01\x00\x00\x00\x05\x48\x65\x6c\x6c\x6f\xAA' | nc localhost "8000"

// makePacket returns a properly formatted packet with the given header and data.
func makePacket(header byte, dataList [][]byte) ([]byte, error) {
    var res []byte
    // Add header.
    res = append(res, header)

    for _, data := range dataList {
        // Get data length and check it's validity.
        dataLen := len(data)
        if dataLen > MAX_DATA_SIZE {
            return nil, errors.New("Data exceeded maximum data size")
        }

        // Grow slice to fit data size and data block.
        res = slices.Grow(res, DATA_BLOCK_SIZE + dataLen)
        // Append length prefix and data.
        binary.BigEndian.PutUint32(res, uint32(dataLen))
        res = append(res, data...)
    }

    return res, nil
}


// dummy is a type-correct function for testing.
func dummy(data [][]byte) ([]byte, error) {
    return nil, nil
}


// echoHandler returns a packet of the appropriate type with the received data.
func echoHandler(data [][]byte) ([]byte, error) {
    return makePacket(ECHO_R, data), nil
}