package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	vim "local/haya14busa/go-vim-server"
)

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
	log.Printf("receive: %#v", msg)
}

func main() {
	flag.Parse()

	cli, closer, err := vim.NewChildClient(&myHandler{})
	if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()

	fmt.Println(cli.Ex("echom 'hi'"))

	cli.Redraw("")
	cli.Redraw("force")

	cli.Normal("gg")

	for i := 0; i < 3; i++ {
		fmt.Println(cli.Expr(fmt.Sprintf("1+%d", i)))
	}

	fmt.Println(cli.Call("matchstr", "testing", "ing"))
	fmt.Println(cli.Call("matchstr", "testing", "ing", 2))
	fmt.Println(cli.Call("matchstr", "testing", "ing", 5))

	if err := cli.Ex("hoge"); err != nil {
		fmt.Printf(":hoge error: %v\n", err)
	}

	if err := cli.Ex("echom 'hi'"); err != nil {
		fmt.Printf(":echom error: %v\n", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		log.Printf("send: %v", scanner.Text())
		if _, err := cli.Conn.Write(scanner.Bytes()); err != nil {
			log.Fatal(err)
		}
	}
}
