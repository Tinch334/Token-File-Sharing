package server


import (
    "fmt"
    "time"
    "context"
    "net"
    "io"
    "encoding/binary"

    "github.com/Tinch334/Token-File-Sharing/internal/constants"
)


// A type for packet type handlers, they take the packets data and return the response that should be sent to the connected host.
// The error string should contain a description of the error.
type TypeHandler func(*Server, [][]byte) ([]byte, error)
// Uses method expressions to pass a reference to the server to handlers whilst keeping the server definition lean.
var handlerMap = map[byte]TypeHandler{
    constants.ERR:      (*Server).dummy,
    constants.USR_CONN: (*Server).connAuthHandler,
    constants.TOK_CONN: (*Server).dummy,
    constants.ECHO_S:   (*Server).echoHandler,
    constants.ECHO_R:   (*Server).dummy,
    constants.FLE_LST:  (*Server).dummy,
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

    fmt.Printf("%v\n", conn.RemoteAddr())

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
    } else if header[0] != constants.USR_CONN {
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


// dummy is a placeholder handler.
func (s *Server) dummy(data [][]byte) ([]byte, error) {
    return nil, nil
}


func (s *Server) connAuthHandler(data [][]byte) ([]byte, error) {
    if len(data) != 2 {
        return makeErrorPacket("Invalid credential format")
    }

    usr := string(data[0])
    psw := string(data[1])

    // Temporary testing setup.
    if usr == "Pepe" && psw == "1234" {
        // Generate new token and send it to host.
        nt := s.th.GenerateToken(usr)
        return makePacket(constants.TOK_RES, [][]byte{[]byte(nt)})
    }

    // Invalid credentials send user and error.
    return makeErrorPacket("Invalid credentials")
}


// echoHandler returns an ECHO_R packet with all the received data.
func (s *Server) echoHandler(data [][]byte) ([]byte, error) {
    return makePacket(constants.ECHO_R, data)
}