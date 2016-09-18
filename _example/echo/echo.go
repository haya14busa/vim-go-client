package main

import (
	"os"

	vim "github.com/haya14busa/vim-go-client"
)

type echoHandler struct{}

func (h *echoHandler) Serve(cli *vim.Client, msg *vim.Message) {
	cli.Send(msg)
}

func main() {
	handler := &echoHandler{}
	cli := vim.NewClient(vim.NewReadWriter(os.Stdin, os.Stdout), handler)
	cli.Start()
}
