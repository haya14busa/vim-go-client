package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	vim "local/haya14busa/go-vim-server"
)

var (
	port       = flag.Int("port", 8765, "The server port")
	servername = flag.String("servername", "VIM", "vim servername")
)

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
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

	cli, err := vim.Connect(addr, *servername, server)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cli.Expr("1+1"))

	log.Println("connected to vim server!")
}
