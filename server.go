package vim

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Message struct {
	MsgID int
	Body  interface{}
}

type Handler interface {
	Serve(io.Writer, *Message)
}

type Server struct {
	Handler Handler // handler to invoke

	conn chan net.Conn
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	srv.conn = make(chan net.Conn, 1)
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

		srv.conn <- conn

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go srv.handleConn(conn)
	}
	return nil
}

func (srv *Server) Connect() net.Conn {
	return <-srv.conn
}

// Serve a new connection.
func (srv *Server) handleConn(c net.Conn) {
	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		msg, err := unmarshalMsg(scanner.Bytes())
		if err != nil {
			// TODO: handler err
			logger.Println(err)
			continue
		}
		srv.Handler.Serve(c, msg)
	}
	if err := scanner.Err(); err != nil {
		logger.Println(err)
	}
}

// unmarshalMsg unmarshals json message from Vim.
// msg format: [{number},{expr}]
func unmarshalMsg(data []byte) (msg *Message, err error) {
	var vimMsg [2]interface{}
	if err := json.Unmarshal(data, &vimMsg); err != nil {
		return nil, fmt.Errorf("fail to decode vim message: %v", err)
	}
	number, ok := vimMsg[0].(float64)
	if !ok {
		return nil, fmt.Errorf("expects message ID, but got %+v", vimMsg[0])
	}
	m := &Message{
		MsgID: int(number),
		Body:  vimMsg[1],
	}
	return m, nil
}
