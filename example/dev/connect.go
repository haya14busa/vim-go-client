package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	vim "local/haya14busa/go-vim-server"
)

var (
	port       = flag.Int("port", 8765, "The server port")
	servername = flag.String("servername", "VIM", "vim servername")
)

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
	log.Printf("receive: %#v", msg)
	if msg.MsgID > 0 && msg.Body == "hi" {
		cli.Write(&vim.Message{msg.MsgID, "hi from connected vim client"})
	}
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

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		log.Printf("send: %v", scanner.Text())
		if _, err := cli.RW.Write(scanner.Bytes()); err != nil {
			log.Fatal(err)
		}
	}
}
