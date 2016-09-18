package main

import (
	"testing"

	vim "github.com/haya14busa/vim-go-client"
)

type myHandler struct{}

func (h *myHandler) Serve(cli *vim.Client, msg *vim.Message) {}

const wantMes = `
Messages maintainer: Bram Moolenaar <Bram@vim.org>
hi!
{'msg': 'hi!'}`

func TestExampleEcho(t *testing.T) {
	cli, closer, err := vim.NewChildClient(&myHandler{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	err = cli.Ex(":message clear")
	err = cli.Ex(":source $GOPATH/src/github.com/haya14busa/vim-go-client/example/echo/echo.vim")
	mes, err := cli.Expr("execute(':message')")
	if err != nil {
		t.Error(err)
	}
	got, ok := mes.(string)
	if !ok {
		t.Fatal("mes should be string")
	}
	if got != wantMes {
		t.Errorf("got:\n%v\nwant:\n%v", got, wantMes)
	}
}
