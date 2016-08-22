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
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handleConn(conn)
	}
}

type Message struct {
	Id   string      `json:"id"`
	Data interface{} `json:"data"`
}

func handleConn(c net.Conn) {
	// Shut down the connection.
	defer c.Close()

	msgID, expr, err := decodeMsg(c)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(msgID, expr)

	// Echo all incoming data.
	// io.Copy(c, c)
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
