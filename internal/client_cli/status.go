package client_cli


type status struct {
	connected bool // Whether there's a connection.
	conn      net.Conn
	token     uin16
}