package vim

import (
	"errors"
	"net"
)

// ErrTimeOut represents time out error.
var ErrTimeOut = errors.New("time out")

// Body represents Message body. e.g. {expr} of `:h ch_sendexpr()`
type Body interface{}

// Message represents rpc message type of JSON channel. `:h channel-use`.
type Message struct {
	MsgID int
	Body  Body
}

// Handler represents go server handler to handle message from Vim.
type Handler interface {
	Serve(*Client, *Message)
}

// Server represents go server.
type Server struct {
	Handler Handler // handler to invoke
}

// Serve starts go server.
func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				logger.Println(err)
				continue
			}
			return err
		}

		cli := NewClient(conn, srv.Handler)

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go cli.Start()
	}
}
