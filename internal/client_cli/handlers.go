package client_cli


import (
    "os"
    "time"
    "net"
)


func (s *status) exitHandler(_ []string) {
    /*
        ¡¡¡Check for connection before disconnecting!!!
    */
    success("Bye")
    os.Exit(0)
}


func (s *status) connectionHandler(subcommand []string) {
    if len(subcommand) == 0 {
        warn("Expected subcommand for `conn`\n")
        return
    }

    switch subcommand[0] {
    case "cn":
        if s.isConnected() {
            warn("Already connected to server\n")
            return
        }

        if len(subcommand) != 4 {
            warn("Invalid format for `conn`\n")
            return
        }

        // Check the message can be created before attempting connection.
        data, err := MakeTokenPacket("", USR_CONN, [][]byte{[]byte{subcommand[2]}, []byte{subcommand[3]}})

        if err != nil {
            warn("Invalid arguments for `conn`\n")
            return
        }

        conn, err := net.DialTimeout("tcp", subcommand[1], time.Seconds(5))

        if err != nil {
            errorPrint("Could not connect to server on `%s`\n", subcommand[1])
            return
        }

        // The connection has been established, update the state.
        s.setConn(conn)
        if err := SendPacket(data, *s.getConn()); err != nil {
            errorPrint("Unable to establish connection with server, closing connection\n")
            if err := s.Close(); err != nil {
                errorPrint("Unable to close connection\n")
            }
            return
        }

        // Await the servers response with a token.
        readData, err := ReadData(*s.getConn())

        if err != nil {
            errorPrint("Unable to read server acknowledgement, closing connection\n")
            if err := s.Close(); err != nil {
                errorPrint("Unable to close connection\n")
            }
            return
        }


    case "dc":
    case "st":
        if !s.isConnected() {
            info("Not connected to server\n")
        } else {
            info("Connected to server with address `%s`\n", (*s.getConn()).RemoteAddr().String())
        }
    }
}