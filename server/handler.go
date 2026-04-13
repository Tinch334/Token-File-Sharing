package server


import (
	"context"
	"net"
	"bufio"
	"strings"
	"fmt"

	"github.com/rs/zerolog/log"
)


// Handles incoming connections to the server.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) (int, bool) {
    reader := bufio.NewReader(conn)
    message, err := reader.ReadString('\n')
    if err != nil {
        log.Printf("Read error: %v", err)
        return 0, false
    }


    ackMsg := strings.ToUpper(strings.TrimSpace(message))
    response := fmt.Sprintf("ACK: %s\n", ackMsg)
    _, err = conn.Write([]byte(response))
    if err != nil {
        log.Printf("Server write error: %v", err)
    }

    return 0, false
}