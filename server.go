package vim

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"
)

var ErrTimeOut = errors.New("time out!")

type Body interface{}

type Message struct {
	MsgID int
	Body  Body
}

type Handler interface {
	Serve(io.Writer, *Message)
}

type Server struct {
	Handler Handler // handler to invoke

	conn      net.Conn
	chConn    chan net.Conn
	responses map[int]chan Body // TODO: need lock?
}

func (srv *Server) initServer() {
	if srv.chConn == nil {
		srv.chConn = make(chan net.Conn, 1)
	}
	if srv.responses == nil {
		srv.responses = make(map[int]chan Body)
	}
}

func (srv *Server) Redraw(force string) error {
	v := []interface{}{"redraw", force}
	conn, err := srv.Connect()
	if err != nil {
		return err
	}
	return json.NewEncoder(conn).Encode(v)
}

func (srv *Server) Ex(cmd string) error {
	var err error
	conn, err := srv.Connect()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(conn)
	err = encoder.Encode([]interface{}{"ex", "let v:errmsg = ''"})
	err = encoder.Encode([]interface{}{"ex", cmd})
	body, err := srv.Expr("v:errmsg")
	if errmsg, ok := body.(string); ok && errmsg != "" {
		err = errors.New(errmsg)
	}
	return err
}

func (srv *Server) Normal(ncmd string) error {
	v := []interface{}{"normal", ncmd}
	conn, err := srv.Connect()
	if err != nil {
		return err
	}
	return json.NewEncoder(conn).Encode(v)
}

func (srv *Server) Expr(expr string) (Body, error) {
	n := srv.prepareResp()
	v := []interface{}{"expr", expr, n}
	conn, err := srv.Connect()
	if err != nil {
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(v); err != nil {
		return nil, err
	}
	return srv.waitResp(n)
}

func (srv *Server) Call(funcname string, args ...interface{}) (Body, error) {
	n := srv.prepareResp()
	v := []interface{}{"call", funcname, args, n}
	conn, err := srv.Connect()
	if err != nil {
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(v); err != nil {
		return nil, err
	}
	return srv.waitResp(n)
}

// prepareResp prepares response from Vim and returns negative number.
// Server.waitResp can wait and get the response.
func (srv *Server) prepareResp() int {
	for {
		n := -int(rand.Int31())
		if _, ok := srv.responses[n]; ok {
			continue
		}
		srv.responses[n] = make(chan Body, 1)
		return n
	}
	return 0
}

// fillResp fills response which is prepared by Server.prepareResp().
func (srv *Server) fillResp(n int, body Body) {
	if ch, ok := srv.responses[n]; ok {
		ch <- body
	}
}

// waitResp waits response which is prepared by Server.prepareResp().
func (srv *Server) waitResp(n int) (Body, error) {
	select {
	case body := <-srv.responses[n]:
		delete(srv.responses, n)
		return body, nil
	case <-time.After(15 * time.Second):
		return nil, ErrTimeOut
	}
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	srv.initServer()
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

		srv.chConn <- conn

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go srv.handleConn(conn)
	}
	return nil
}

// Connect returns connection to vim. If connection hasn't been established
// yet, wait for connection establishment.
func (srv *Server) Connect() (net.Conn, error) {
	if srv.conn != nil {
		return srv.conn, nil
	}
	srv.initServer()
	logger.Println("Connect() waits connection", srv.chConn)
	select {
	case srv.conn = <-srv.chConn:
		return srv.conn, nil
	case <-time.After(15 * time.Second):
		return nil, ErrTimeOut
	}
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
		if msg.MsgID < 0 {
			srv.fillResp(msg.MsgID, msg.Body)
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
