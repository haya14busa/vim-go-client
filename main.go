package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	port = flag.Int("port", 8765, "The server port")
)

var vimCon = make(chan net.Conn, 1)

func main() {
	flag.Parse()

	addr := fmt.Sprintf("localhost:%d", *port)

	// Listen on TCP port *port on all interfaces.
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	go serve(l)

	p, err := NewVimServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	scanner := bufio.NewScanner(os.Stdin)
	con := <-vimCon
	defer con.Close()
	log.Println("connected to vim server!")
	for scanner.Scan() {
		log.Printf("send: %v", scanner.Text())
		if _, err := con.Write(scanner.Bytes()); err != nil {
			log.Println(err)
		}
	}
}

func serve(l net.Listener) {
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
	// defer c.Close()
	vimCon <- c

	fmt.Printf("local addr: %v\n", c.LocalAddr())
	fmt.Printf("remote addr: %v\n", c.RemoteAddr())

	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		log.Printf("receive: %v", scanner.Text())
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
