package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	vim "local/haya14busa/go-vim-server"
)

type myHandler struct{}

func (h *myHandler) Serve(w io.Writer, msg *vim.Message) {
	log.Printf("receive: %#v", msg)
}

func main() {
	flag.Parse()

	addr := fmt.Sprintf("localhost:0")

	// Listen on TCP port *port on all interfaces.
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	server := &vim.Server{Handler: &myHandler{}}
	go server.Serve(l)

	p, err := vim.NewChildVimServer(l.Addr().String())
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	server.Ex("echom 'hi'")
	server.Redraw("")
	server.Redraw("force")

	server.Normal("gg")

	for i := 0; i < 3; i++ {
		fmt.Println(server.Expr(fmt.Sprintf("1+%d", i)))
	}

	fmt.Println(server.Call("matchstr", "testing", "ing"))
	fmt.Println(server.Call("matchstr", "testing", "ing", 2))
	fmt.Println(server.Call("matchstr", "testing", "ing", 5))

	if err := server.Ex("hoge"); err != nil {
		fmt.Printf(":hoge error: %v\n", err)
	}

	if err := server.Ex("echom 'hi'"); err != nil {
		fmt.Printf(":echom error: %v\n", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	conn, err := server.Connect()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to vim server!")
	for scanner.Scan() {
		log.Printf("send: %v", scanner.Text())
		if _, err := conn.Write(scanner.Bytes()); err != nil {
			log.Println(err)
		}
	}
}
