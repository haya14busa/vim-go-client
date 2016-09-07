package vim_test

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	vim "github.com/haya14busa/vim-go-client"
)

var cli *vim.Client

var defaultServeFunc = func(cli *vim.Client, msg *vim.Message) {}

var (
	serveFunc  = defaultServeFunc
	serveFunMu sync.RWMutex
)

var vimArgs = []string{"-Nu", "NONE", "-i", "NONE", "-n"}

var waitLog = func() { time.Sleep(1 * time.Millisecond) }

type testHandler struct {
	f func(cli *vim.Client, msg *vim.Message)
}

func (h *testHandler) Serve(cli *vim.Client, msg *vim.Message) {
	fn := h.f
	if fn == nil {
		fn = defaultServeFunc
	}
	fn(cli, msg)
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
	var serveFuncCalledMu sync.RWMutex

	serveFunMu.Lock()
	serveFunc = func(cli *vim.Client, msg *vim.Message) {
		// t.Log(msg)
		serveFuncCalledMu.Lock()
		serveFuncCalled = true
		serveFuncCalledMu.Unlock()
	}
	defer func() {
		serveFunMu.Unlock()
		serveFunc = defaultServeFunc
	}()

	cli, closer, err := vim.NewChildClient(&testHandler{f: serveFunc}, vimArgs)
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()
	if _, err = cli.Expr("1 + 1"); err != nil {
		t.Fatal(err)
	}

	serveFuncCalledMu.Lock()
	if !serveFuncCalled {
		t.Error("serveFunc must be called")
	}
	serveFuncCalledMu.Unlock()

	status, err := cli.Expr("ch_status(g:vim_go_client_handler)")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%s", status) != "open" {
		t.Error("channel must be open")
	}
}

func TestClient_Send(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Send(&vim.Message{MsgID: -1, Body: "hi, how are you?"})
	waitLog()
	if !containsString(tmp, "hi, how are you?") {
		t.Error("cli.Send should send message to Vim")
	}
}

func TestClient_Write(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Write([]byte("hi, how are you?"))
	waitLog()
	if !containsString(tmp, "hi, how are you?") {
		t.Error("cli.Write should send message to Vim")
	}
}

func TestClient_Redraw_force(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Redraw("force")
	waitLog()
	if !containsString(tmp, ": redraw") {
		t.Error(`cli.Redraw("force") should redraw Vim`)
	}
}

func TestClient_Redraw(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Redraw("")
	waitLog()
	if !containsString(tmp, ": redraw") {
		t.Error(`cli.Redraw("") should redraw`)
	}
}

func TestClient_Ex(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Ex("echo 'hi'")
	waitLog()
	if !containsString(tmp, ": Executing ex command") {
		t.Error(`cli.Ex() should execute ex command`)
	}
}

func TestClient_Normal(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Normal("gg")
	waitLog()
	if !containsString(tmp, ": Executing normal command") {
		t.Error(`cli.Normal() should execute normal command`)
	}
}

func TestClient_Expr_log(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Expr("0")
	waitLog()
	if !containsString(tmp, ": Evaluating expression") {
		t.Error(`cli.Expr() should evaluate expr`)
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

func TestClient_Call(t *testing.T) {
	tmp, err := ioutil.TempFile("", "vim-go-client-test-log")
	if err != nil {
		t.Fatal(err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())
	cli.Expr(fmt.Sprintf("ch_logfile('%s', 'w')", tmp.Name()))
	cli.Call("eval", `"1"`)
	waitLog()
	if !containsString(tmp, ": Calling") {
		t.Error(`cli.Expr() should call func`)
	}
}

func TestClient_Call_resp(t *testing.T) {
	got, err := cli.Call("eval", `v:true`)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, true) {
		t.Error("cli.Call() should return responses")
	}
}

// containsString checks reader contains str. It doens't handle NL and consumes reader!
func containsString(r io.Reader, str string) bool {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), str) {
			return true
		}
	}
	return false
}
