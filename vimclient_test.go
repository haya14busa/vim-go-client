package vim_test

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	vim "github.com/haya14busa/vim-go-client"
)

var cli *vim.Client

var defaultServeFunc = func(cli *vim.Client, msg *vim.Message) {}
var serveFunc = defaultServeFunc

var vimArgs = []string{"-Nu", "NONE", "-i", "NONE", "-n"}

type testHandler struct{}

func (h *testHandler) Serve(cli *vim.Client, msg *vim.Message) {
	serveFunc(cli, msg)
}

func TestMain(m *testing.M) {
	c, closer, err := vim.NewChildClient(&testHandler{}, vimArgs)
	if err != nil {
		log.Fatal(err)
	}
	cli = c
	code := m.Run()
	closer.Close()
	os.Exit(code)
}

func BenchmarkNewChildClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, closer, err := vim.NewChildClient(&testHandler{}, vimArgs)
		if err != nil {
			b.Fatal(err)
		}
		defer closer.Close()
	}
}

func TestNewChildClient(t *testing.T) {
	serveFuncCalled := false
	serveFunc = func(cli *vim.Client, msg *vim.Message) {
		// t.Log(msg)
		serveFuncCalled = true
	}
	defer func() { serveFunc = defaultServeFunc }()

	cli, closer, err := vim.NewChildClient(&testHandler{}, vimArgs)
	if err != nil {
		t.Fatal(err)
	}
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

func TestClient_Expr(t *testing.T) {
	tests := []struct {
		in   string
		want interface{}
	}{
		{in: "1 + 1", want: float64(2)},
		{in: `'hello ' . 'world!'`, want: "hello world!"},
		{in: "{}", want: map[string]interface{}{}},
		{in: "[1,2,3]", want: []interface{}{float64(1), float64(2), float64(3)}},
		{in: "v:false", want: false},
		{in: "v:true", want: true},
		{in: "v:none", want: nil},
		{in: "v:null", want: nil},
		{in: "v:count", want: float64(0)},
		{in: "{x -> x * x}(2)", want: float64(4)},
	}

	for _, tt := range tests {
		got, err := cli.Expr(tt.in)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("cli.Expr(%v) == %#v, want %#v", tt.in, got, tt.want)
		}
	}
}

func TestClient_Expr_fail(t *testing.T) {
	tests := []struct {
		in string
	}{
		{in: ""},
		{in: "1 + {}"},
		{in: "xxx"},
		{in: "{x -> x * x}"},
		{in: "function('tr')"},
	}

	for _, tt := range tests {
		got, err := cli.Expr(tt.in)
		if err != nil {
			if err != vim.ErrExpr {
				t.Errorf("cli.Expr(%v) want vim.ErrExpr, got %v", tt.in, err)
			}
		} else {
			t.Errorf("cli.Expr(%v) expects error but got %v", tt.in, got)
		}
	}
}
