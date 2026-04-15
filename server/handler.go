package server


import (
    "fmt"
    "slices"
    "errors"
    "time"
    "context"
    "net"
    "io"
    "encoding/binary"
)


// A type for packet type handlers, they take the packets data and return the response that should be sent to the connected host.
// The error string should contain a description of the error.
type TypeHandler func(*Server, [][]byte) ([]byte, error)
// Uses method expressions to pass a reference to the server to handlers whilst keeping the server definition lean.
var handlerMap = map[byte]TypeHandler{
    ERR:      (*Server).dummy,
    USR_CONN: (*Server).connAuthHandler,
    TOK_CONN: (*Server).dummy,
    ECHO_S:   (*Server).echoHandler,
    ECHO_R:   (*Server).dummy,
    FLE_LST:  (*Server).dummy,
}

/* Token request with "Pepe", "1234":
    printf '\x01\x00\x00\x00\x00\x00\x04\x50\x65\x70\x65\xCC\x00\x00\x00\x04\x31\x32\x33\x34\xAA' | nc localhost "8000"

Echo with "Hello":
    printf '\x06\x00\x04\xxx\xxx\xxx\xxx\x00\x00\x00\x05\x48\x65\x6c\x6c\x6f\xAA' | nc localhost "8000"
*/

  

// Handles incoming connections to the server.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) (int, error) {
    defer conn.Close()

    // Set read deadline in case connected host disconnects.
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

    // Indicates weather an auth token was present, a nil check on authBlock could be used; A dedicated variable is cleaner.
    var authPresent bool
    // Stores authentication data from auth block.
    var authString string

    // Read packet type.
    header := make([]byte, 1)
    if _, err := io.ReadFull(conn, header); err != nil {
        if err == io.ErrUnexpectedEOF {
            sendError("Truncated header", conn)
        }
        return 0, fmt.Errorf("Error reading header: %w", err)
    }

    // Read auth data size.
    authSizeBytes := make([]byte, 2)
    if _, err := io.ReadFull(conn, authSizeBytes); err != nil {
        if err == io.ErrUnexpectedEOF {
            sendError("Truncated auth size", conn)
        }
        return 0, fmt.Errorf("Error reading auth size: %w", err)
    }

    // Convert auth data size and check if there's an auth token.
    authSize := binary.BigEndian.Uint16(authSizeBytes)
    if authSize == 0 {
        authPresent = false
    } else {
        authBlock := make([]byte, authSize)
        if _, err := io.ReadFull(conn, authBlock); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated auth data", conn)
            }
            return 0, fmt.Errorf("Error reading auth data: %w", err)
        }

        authPresent = true
        authString = string(authBlock)
    }

    // Check if the packet has authentication data.
    if authPresent {
        // Check if token is valid, if so reset it's timer, otherwise send and error to the host and close the connection.
        if s.th.ValidateToken(authString) {
            s.th.ResetTokenTimer(authString)
        } else {
            sendError("Invalid authentication", conn)
            return 0, fmt.Errorf("Invalid auth, with token: %s\n", authString)
        }
    } else if header[0] != USR_CONN {
        sendError(fmt.Sprintf("Operation requires authentication: %X", header[0]), conn)
        return 0, fmt.Errorf("Operation without auth: %X\n", header[0])
    }

    // Read packet data.
    data, err := readData(conn)
    if err != nil {
        return 0, err
    }

    // Dispatch to appropriate handler.
    handler, ok := handlerMap[header[0]]
    if !ok {
        return 0, sendError(fmt.Sprintf("Unknown command: %X", header[0]), conn)
    }

    // Call handler with received data and send response.
    resData, err := handler(s, data)
    if err != nil {
        return 0, sendError(fmt.Sprintf("Error handling %X, encountered: %v", header[0], err), conn)
    }

    return 0, sendPacket(resData, conn)
}


// readData reads all of the packet's data blocks.
func readData(conn net.Conn) ([][]byte, error) {
    // Read data blocks until ADDITIONAL_BLOCK_NO.
    var data [][]byte
    for {
        // Read packet data size.
        packetSizeBytes := make([]byte, 4)
        if _, err := io.ReadFull(conn, packetSizeBytes); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated block size", conn)
            }
            return nil, fmt.Errorf("Error reading block size: %w", err)
        }

        // Read packet data.
        blockSize := binary.BigEndian.Uint32(packetSizeBytes)
        block := make([]byte, blockSize)
        if _, err := io.ReadFull(conn, block); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated block data", conn)
            }
            return nil, fmt.Errorf("Error reading block data: %w", err)
        }

        // Store read packet data.
        data = append(data, block)

        // Read continuation indicator.
        cIndicator := make([]byte, 1)
        if _, err := io.ReadFull(conn, cIndicator); err != nil {
            if err == io.ErrUnexpectedEOF {
                sendError("Truncated continuation data", conn)
            }
            return nil, fmt.Errorf("Error continuation bytes: %w", err)
        }

        fmt.Printf("data:%X  blockSize:%d  block:%X  cont:%X\n",
            data, blockSize, block, cIndicator)

        // Check if there's more data to be read.
        if cIndicator[0] == ADDITIONAL_BLOCK_NO {
            break
        }
    }

    return data, nil
}

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
func (s *Server) dummy(data [][]byte) ([]byte, error) {
    return nil, nil
}


func (s *Server) connAuthHandler(data [][]byte) ([]byte, error) {
    usr := string(data[0])
    psw := string(data[1])

    // Temporary testing setup.
    if usr == "Pepe" && psw == "1234" {
        // Generate new token and send it to host.
        nt := s.th.GenerateToken(usr)
        return makePacket(TOK_RES, [][]byte{[]byte(nt)})
    }

    // Invalid credentials send user and error.
    return makePacket(ERR, [][]byte{[]byte("Invalid credentials")})
}


// echoHandler returns an ECHO_R packet with all the received data.
func (s *Server) echoHandler(data [][]byte) ([]byte, error) {
    return makePacket(ECHO_R, data)
}