package main

import (
	"fmt"
	"log"
	"os"
	"time"

	vim "local/haya14busa/go-vim-server"
)

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
	log.Printf("receive: %#v", msg)
	if msg.Body == "hi" {
		cli.Ex("echom 'hi, how are you?'")
	}
	if msg.MsgID > 0 {
		start := time.Now()
		log.Println(cli.Expr("eval(join(range(10), '+'))"))
		log.Printf("cli.Expr: finished in %v", time.Now().Sub(start))
	}
}

func main() {
	handler := &myHandler{}
	cli := vim.NewClient(vim.NewReadWriter(os.Stdin, os.Stdout), handler)
	done := make(chan error, 1)
	go func() {
		done <- cli.Start()
	}()

	cli.Ex("echom 'hi'")
	log.Println(cli.Expr("1+1"))

	select {
	case err := <-done:
		fmt.Printf("exit with error: %v\n", err)
		fmt.Println("bye;)")
	}
}
