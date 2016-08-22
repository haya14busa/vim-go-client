package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	port = flag.Int("port", 8765, "The server port")
)

func main() {
	flag.Parse()
	// Listen on TCP port *port on all interfaces.
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		c, ok := conn.(*net.TCPConn)
		if !ok {
			log.Println(err)
			continue
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		// go handleConn(conn)
		go handleConn(c)
	}
}

type Message struct {
	Id   string      `json:"id"`
	Data interface{} `json:"data"`
}

func handleConn(c net.Conn) {
	// XXX: Do not shutdown connection for vim client. is it ok???
	// defer c.Close()

	msgID, expr, err := decodeMsg(c)
	if err != nil {
		log.Println(err)
		return
	}

	resp := []interface{}{msgID, expr}

	e := json.NewEncoder(c)
	if err := e.Encode(resp); err != nil {
		log.Println(err)
	}
}

// decodeMsg decodes message from vim channel.
func decodeMsg(r io.Reader) (msgID int, expr interface{}, err error) {
	d := json.NewDecoder(r)
	var vimMsg interface{} // [{number},{expr}]
	if err := d.Decode(&vimMsg); err != nil {
		return 0, nil, fmt.Errorf("fail to decode vim message: %v", err)
	}
	ms, ok := vimMsg.([]interface{})
	if !ok {
		return 0, nil, fmt.Errorf("expects [{number},{expr}], but got %v", vimMsg)
	}
	number, ok := ms[0].(float64)
	if !ok {
		return 0, nil, fmt.Errorf("expects message ID, but got %+v", ms[0])
	}
	msgID = int(number)
	expr = ms[1]
	return msgID, expr, nil
}
