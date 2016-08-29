package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	vim "github.com/haya14busa/vim-go-client"
)

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {
	log.Printf("receive: %#v", msg)
	if msg.MsgID > 0 {
		start := time.Now()
		log.Println(cli.Expr("eval(join(range(10), '+'))"))
		log.Printf("cli.Expr: finished in %v", time.Now().Sub(start))
	}
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

	{
		start := time.Now()
		for i := 0; i < 3; i++ {
			fmt.Println(cli.Expr(fmt.Sprintf("1+%d", i)))
		}
		log.Printf("cli.Expr * 3: finished in %v", time.Now().Sub(start))
	}

	{
		start := time.Now()
		fmt.Println(cli.Call("matchstr", "testing", "ing"))
		fmt.Println(cli.Call("matchstr", "testing", "ing", 2))
		fmt.Println(cli.Call("matchstr", "testing", "ing", 5))
		log.Printf("cli.Call: finished in %v", time.Now().Sub(start))
	}

	if err := cli.Ex("hoge"); err != nil {
		fmt.Printf(":hoge error: %v\n", err)
	}

	if err := cli.Ex("echom 'hi'"); err != nil {
		fmt.Printf(":echom error: %v\n", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		log.Printf("send: %v", scanner.Text())
		if _, err := cli.Write(scanner.Bytes()); err != nil {
			log.Fatal(err)
		}
	}
}
