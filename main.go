package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	port = flag.Int("port", 8765, "The server port")
)

type myHandler struct{}

func (h *myHandler) Serve(w io.Writer, msg *Message) {
	log.Printf("receive: %#v", msg)
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf("localhost:%d", *port)

	// Listen on TCP port *port on all interfaces.
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	server := &Server{Handler: &myHandler{}}
	go server.Serve(l)

	p, err := NewVimServer(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	scanner := bufio.NewScanner(os.Stdin)
	conn := server.Connect()
	defer conn.Close()
	log.Println("connected to vim server!")
	for scanner.Scan() {
		log.Printf("send: %v", scanner.Text())
		if _, err := conn.Write(scanner.Bytes()); err != nil {
			log.Println(err)
		}
	}
}
