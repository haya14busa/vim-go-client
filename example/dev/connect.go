package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	vim "local/haya14busa/go-vim-server"
)

var (
	port = flag.Int("port", 8765, "The server port")
)

type myHandler struct{}

func (h *myHandler) Serve(w io.Writer, msg *vim.Message) {
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
	server := &vim.Server{Handler: &myHandler{}}
	go server.Serve(l)

	conn, err := server.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("connected to vim server!")
}
