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

type Client struct {
	RW      io.ReadWriter
	handler Handler

	responses map[int]chan Body // TODO: need lock?
}

type ChildCliCloser struct {
	listener net.Listener
	process  *Process
}

func (c *ChildCliCloser) Close() error {
	var err error
	err = c.process.Close()
	err = c.listener.Close()
	return err
}

var _ Handler = &getCliHandler{}

type getCliHandler struct {
	handler Handler

	connected bool
	chCli     chan *Client
}

func (h *getCliHandler) Serve(cli *Client, msg *Message) {
	if !h.connected {
		h.chCli <- cli
		h.connected = true
	}
	h.handler.Serve(cli, msg)
}

func NewClient(rw io.ReadWriter, handler Handler) *Client {
	return &Client{
		RW:      rw,
		handler: handler,

		responses: make(map[int]chan Body),
	}
}

func NewChildClient(handler Handler) (*Client, *ChildCliCloser, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, err
	}

	p, err := NewChildVimServer(l.Addr().String())
	if err != nil {
		return nil, nil, err
	}

	h := &getCliHandler{
		handler: handler,
		chCli:   make(chan *Client, 1),
	}

	server := &Server{Handler: h}
	go server.Serve(l)

	closer := &ChildCliCloser{
		listener: l,
		process:  p,
	}

	select {
	case cli := <-h.chCli:
		return cli, closer, nil
	case <-time.After(15 * time.Second):
		closer.Close()
		return nil, nil, ErrTimeOut
	}
}

func (cli *Client) Write(msg *Message) error {
	v := [2]interface{}{msg.MsgID, msg.Body}
	return json.NewEncoder(cli.RW).Encode(v)
}

func (cli *Client) Redraw(force string) error {
	v := []interface{}{"redraw", force}
	return json.NewEncoder(cli.RW).Encode(v)
}

func (cli *Client) Ex(cmd string) error {
	var err error
	encoder := json.NewEncoder(cli.RW)
	err = encoder.Encode([]interface{}{"ex", "let v:errmsg = ''"})
	err = encoder.Encode([]interface{}{"ex", cmd})
	body, err := cli.Expr("v:errmsg")
	if errmsg, ok := body.(string); ok && errmsg != "" {
		err = errors.New(errmsg)
	}
	return err
}

func (cli *Client) Normal(ncmd string) error {
	v := []interface{}{"normal", ncmd}
	return json.NewEncoder(cli.RW).Encode(v)
}

func (cli *Client) Expr(expr string) (Body, error) {
	n := cli.prepareResp()
	v := []interface{}{"expr", expr, n}
	if err := json.NewEncoder(cli.RW).Encode(v); err != nil {
		return nil, err
	}
	return cli.waitResp(n)
}

func (cli *Client) Call(funcname string, args ...interface{}) (Body, error) {
	n := cli.prepareResp()
	v := []interface{}{"call", funcname, args, n}
	if err := json.NewEncoder(cli.RW).Encode(v); err != nil {
		return nil, err
	}
	return cli.waitResp(n)
}

// prepareResp prepares response from Vim and returns negative number.
// Server.waitResp can wait and get the response.
func (cli *Client) prepareResp() int {
	if cli.responses == nil {
		cli.responses = make(map[int]chan Body)
	}
	for {
		n := -int(rand.Int31())
		if _, ok := cli.responses[n]; ok {
			continue
		}
		cli.responses[n] = make(chan Body, 1)
		return n
	}
	return 0
}

// fillResp fills response which is prepared by Server.prepareResp().
func (cli *Client) fillResp(n int, body Body) {
	if ch, ok := cli.responses[n]; ok {
		ch <- body
	}
}

// waitResp waits response which is prepared by Server.prepareResp().
func (cli *Client) waitResp(n int) (Body, error) {
	select {
	case body := <-cli.responses[n]:
		delete(cli.responses, n)
		return body, nil
	case <-time.After(15 * time.Second):
		return nil, ErrTimeOut
	}
}

// Serve a new connection.
func (cli *Client) Start() error {
	scanner := bufio.NewScanner(cli.RW)
	for scanner.Scan() {
		msg, err := unmarshalMsg(scanner.Bytes())
		if err != nil {
			// TODO: handler err
			logger.Println(err)
			continue
		}
		if msg.MsgID < 0 {
			cli.fillResp(msg.MsgID, msg.Body)
		}
		go cli.handler.Serve(cli, msg)
	}
	if err := scanner.Err(); err != nil {
		logger.Println(err)
		return err
	}
	return nil
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
