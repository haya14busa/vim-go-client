// Package vim provides Vim client and server implementations.
// You can start Vim as a server as a child process or connect to Vim, and
// communicate with it via TCP or stdin/stdout.
// :h channel.txt
package vim

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"
)

// ErrExpr represents "expr" command error.
var ErrExpr = errors.New("the evaluation fails or the result can't be encoded in JSON")

// Client represents Vim client.
type Client struct {
	// rw is readwriter for communication between go-server and vim-server.
	rw io.ReadWriter

	// handler handles message from Vim.
	handler Handler

	// responses handles response from Vim.
	responses map[int]chan Body // TODO: need lock?
}

// ChildCliCloser is closer of child Vim client process.
type ChildCliCloser struct {
	listener net.Listener
	process  *Process
}

// Close closes child Vim client process.
func (c *ChildCliCloser) Close() error {
	var err error
	err = c.process.Close()
	err = c.listener.Close()
	return err
}

var _ Handler = &getCliHandler{}

// getCliHandler is handler to get one connected Vim client.
type getCliHandler struct {
	handler Handler

	connected   bool
	connectedMu sync.RWMutex
	chCli       chan *Client
}

func (h *getCliHandler) Serve(cli *Client, msg *Message) {
	h.connectedMu.Lock()
	if !h.connected {
		h.chCli <- cli
		h.connected = true
	}
	h.connectedMu.Unlock()
	h.handler.Serve(cli, msg)
}

// NewClient creates Vim client.
func NewClient(rw io.ReadWriter, handler Handler) *Client {
	return &Client{
		rw:      rw,
		handler: handler,

		responses: make(map[int]chan Body),
	}
}

// NewChildClient creates connected child process Vim client.
func NewChildClient(handler Handler, args []string) (*Client, *ChildCliCloser, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, nil, err
	}

	p, err := NewChildVimServer(l.Addr().String(), args)
	if err != nil {
		return nil, nil, err
	}

	h := &getCliHandler{
		handler: handler,
		chCli:   make(chan *Client),
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

// Send sends message to Vim.
func (cli *Client) Send(msg *Message) error {
	v := [2]interface{}{msg.MsgID, msg.Body}
	return json.NewEncoder(cli.rw).Encode(v)
}

// Write writes raw message to Vim.
func (cli *Client) Write(p []byte) (n int, err error) {
	return cli.rw.Write(p)
}

// Redraw runs command "redraw" (:h channel-commands).
func (cli *Client) Redraw(force string) error {
	v := []interface{}{"redraw", force}
	return json.NewEncoder(cli.rw).Encode(v)
}

// Ex runs command "ex" (:h channel-commands).
func (cli *Client) Ex(cmd string) error {
	var err error
	encoder := json.NewEncoder(cli.rw)
	err = encoder.Encode([]interface{}{"ex", cmd})
	return err
}

// Normal runs command "normal" (:h channel-commands).
func (cli *Client) Normal(ncmd string) error {
	v := []interface{}{"normal", ncmd}
	return json.NewEncoder(cli.rw).Encode(v)
}

// Expr runs command "expr" (:h channel-commands).
func (cli *Client) Expr(expr string) (Body, error) {
	n := cli.prepareResp()
	v := []interface{}{"expr", expr, n}
	if err := json.NewEncoder(cli.rw).Encode(v); err != nil {
		return nil, err
	}
	body, err := cli.waitResp(n)
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%s", body) == "ERROR" {
		return nil, ErrExpr
	}
	return body, nil
}

// Call runs command "call" (:h channel-commands).
func (cli *Client) Call(funcname string, args ...interface{}) (Body, error) {
	n := cli.prepareResp()
	v := []interface{}{"call", funcname, args, n}
	if err := json.NewEncoder(cli.rw).Encode(v); err != nil {
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

// Start starts to wait message from Vim.
func (cli *Client) Start() error {
	scanner := bufio.NewScanner(cli.rw)
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
