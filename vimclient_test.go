package vim_test

import (
	"fmt"
	"testing"

	vim "github.com/haya14busa/vim-go-client"
)

var defaultServeFunc = func(cli *vim.Client, msg *vim.Message) {}
var serveFunc = defaultServeFunc

type testHandler struct{}

func (h *testHandler) Serve(cli *vim.Client, msg *vim.Message) {
	serveFunc(cli, msg)
}

func TestNewChildClient(t *testing.T) {
	serveFuncCalled := false
	serveFunc = func(cli *vim.Client, msg *vim.Message) {
		// t.Log(msg)
		serveFuncCalled = true
	}
	defer func() { serveFunc = defaultServeFunc }()

	cli, closer, err := vim.NewChildClient(&testHandler{})
	defer closer.Close()
	if _, err = cli.Expr("1 + 1"); err != nil {
		t.Fatal(err)
	}
	if !serveFuncCalled {
		t.Error("serveFunc must be called")
	}
	status, err := cli.Expr("ch_status(g:vim_go_client_handler)")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%s", status) != "open" {
		t.Error("channel must be open")
	}
}
