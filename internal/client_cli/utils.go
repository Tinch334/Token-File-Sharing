package client_cli


import (
    "net"
)


type status struct {
    conn      *net.Conn // Should be null when there's not a connection.
    token     string
}

// newStatus returns a new status for a client that has just been initialized.
func newStatus() *status {
    st := status {
        conn:  nil,
        token: "",
    }

    return &st
}

// isConnected returns "true" if the status's connection is not "nil".
func (s *status) isConnected() bool {
    return s.conn != nil
}

// getConn returns the connection.
func (s *status) getConn() *net.Conn {
    return s.conn
}

// setConn sets the connection.
func (s *status) setConn(nc *net.Conn) {
    s.conn = nc
}


// A type for command handlers, they should be self contained, handling errors and side effects internally.
// The error string should contain a description of the error.
type TypeHandler func(*status, []string) ()

// dummy is a placeholder handler.
func (s *status) dummy(subcommands []string) () {
    return
}
