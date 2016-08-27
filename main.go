package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
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

func handleConn(c net.Conn) {
	defer c.Close()

	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		msgID, expr, err := unmarshalMsg(scanner.Bytes())
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println(msgID, expr)
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

// unmarshalMsg unmarshals json message from Vim.
// msg format: [{number},{expr}]
func unmarshalMsg(data []byte) (msgID int, expr interface{}, err error) {
	var vimMsg [2]interface{}
	if err := json.Unmarshal(data, &vimMsg); err != nil {
		return 0, nil, fmt.Errorf("fail to decode vim message: %v", err)
	}
	number, ok := vimMsg[0].(float64)
	if !ok {
		return 0, nil, fmt.Errorf("expects message ID, but got %+v", vimMsg[0])
	}
	msgID = int(number)
	expr = vimMsg[1]
	return msgID, expr, nil
}
