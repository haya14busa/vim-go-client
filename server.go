package vim

import (
	"errors"
	"net"
)

var ErrTimeOut = errors.New("time out!")

type Body interface{}

type Message struct {
	MsgID int
	Body  Body
}

type Handler interface {
	Serve(*Client, *Message)
}

type Server struct {
	Handler Handler // handler to invoke
}

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
		go cli.handleConn()
	}
	return nil
}
