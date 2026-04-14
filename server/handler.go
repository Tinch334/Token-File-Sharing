package server


import (
    "errors"
    "time"
    "slices"
    "context"
    "net"
    "io"
    "encoding/binary"

    "fmt"

	//"github.com/rs/zerolog/log"
)


// A type for packet type handlers, they take the packets data and return the response that should be sent to the connected host.
// The error string should contain a description of the error.
type TypeHandler func([][]byte) ([]byte, error)

var handlerMap = map[byte]TypeHandler{
    ERR:      dummy,
    USR_CONN: dummy,
    TOK_CONN: dummy,
    ECHO_S:   echoHandler,
    ECHO_R:   dummy,
    FLE_LST:  dummy,
}


// Handles incoming connections to the server.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) (int, error) {
    defer conn.Close()

    // Set read deadline in case connected host disconnects.
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

    // Read packet type.
    header := make([]byte, 1)
    if _, err := io.ReadFull(conn, header); err != nil {
        if err == io.ErrUnexpectedEOF {
            sendError("Truncated header", conn)
        }
        return 0, fmt.Errorf("Error reading header: %w", err)
    }

    // Read data blocks until ADDITIONAL_BLOCK_NO.
    var data [][]byte
    for {
        // Read packet data size.
        packetSizeBytes := make([]byte, 4)
        if _, err := io.ReadFull(conn, packetSizeBytes); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated block size", conn)
            }
            return 0, fmt.Errorf("Error reading block size: %w", err)
        }

        // Read packet data.
        blockSize := binary.BigEndian.Uint32(packetSizeBytes)
        block := make([]byte, blockSize)
        if _, err := io.ReadFull(conn, block); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated block data", conn)
            }
            return 0, fmt.Errorf("Error reading block data: %w", err)
        }

        // Store read packet data.
        data = append(data, block)

        // Read continuation indicator.
        cIndicator := make([]byte, 1)
        if _, err := io.ReadFull(conn, cIndicator); err != nil {
            // Invalid packet size.
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated continuation data", conn)
            }
            return 0, fmt.Errorf("Error continuation bytes: %w", err)
        }

        fmt.Printf("header:%X  blockSize:%d  block:%X  cont:%X\n",
            header, blockSize, block, cIndicator)

        // Check if there's more data to be read.
        if cIndicator[0] == ADDITIONAL_BLOCK_NO {
            break
        }
    }

    // Dispatch to appropriate handler.
    handler, ok := handlerMap[header[0]]
    if !ok {
        return 0, sendError(fmt.Sprintf("Unknown command: %X", header[0]), conn)
    }

    fmt.Printf("Data:%X\n", data)

    // Call handler with received data and send response.
    resData, err := handler(data)
    if err != nil {
        return 0, sendError(fmt.Sprintf("Error handling %X, encountered: %v", header[0], err), conn)
    }

    return 0, sendPacket(resData, conn)
}

// printf '\x04\x00\x00\x00\x05\x48\x65\x6c\x6c\x6f\xAA' | nc localhost "8000"

// makePacket returns a properly formatted packet with the given header and data.
func makePacket(header byte, dataList [][]byte) ([]byte, error) {
    var res []byte
    // Add header.
    res = []byte{header}

    for i, block := range dataList {
        // Get data length and check it's validity.
        if len(block) > MAX_DATA_SIZE {
            return nil, errors.New("Data exceeded maximum data size")
        }

        lenBuf := make([]byte, 4)
        binary.BigEndian.PutUint32(lenBuf, uint32(len(block)))

        // Grow slice to fit block size, block and continuation indicator, append data.
        res = slices.Grow(res, DATA_BLOCK_SIZE + len(block) + 1)
        res = append(res, lenBuf...)
        res = append(res, block...)
    
        // Set continuation byte.
        if i < len(dataList) - 1 {
            res = append(res, ADDITIONAL_BLOCK_YES)
        } else {
            res = append(res, ADDITIONAL_BLOCK_NO)
        }
    }

    return res, nil
}


// sendPacket sends the given data to the connection.
func sendPacket(data []byte, conn net.Conn) error {
    _, err := conn.Write(data)
    return err
}


// sendError sends an ERR packet to conn with the given error string, the error will not be returned.
// An error is only returned in case of a send error.
func sendError(errorMsg string, conn net.Conn) error {
    // Create a slice of slices with the error message.
    packet, err := makePacket(ERR, [][]byte{[]byte(errorMsg)})
    if err != nil {
        return fmt.Errorf("Error building error packet: %w\n", err)
    }

    if  _, err = conn.Write(packet); err != nil {
        return fmt.Errorf("Error building error packet: %w\n", err)
    }

    return err
}


// dummy is a placeholder handler.
func dummy(data [][]byte) ([]byte, error) {
    return nil, nil
}


// echoHandler returns an ECHO_R packet with all the received data.
func echoHandler(data [][]byte) ([]byte, error) {
    return makePacket(ECHO_R, data)
}