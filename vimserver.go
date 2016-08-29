package vim

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/haya14busa/vim-go-client/remote"
)

// Process represents Vim server process.
type Process struct {
	cmd  *exec.Cmd
	done chan error

	script *os.File
}

// Close closes Vim process.
func (p *Process) Close() error {
	os.Remove(p.script.Name())
	return p.cmd.Process.Signal(os.Interrupt)
}

// send [0, "init connection"] to go server to get initial connection.
const connectScript = `
call ch_logfile('/tmp/vimchannellog', 'w')
while 1
	let g:vim_go_client_handler = ch_open('{{ .Addr }}')
	if ch_status(g:vim_go_client_handler) is# 'open'
		call ch_sendraw(g:vim_go_client_handler, "[0, \"init connection\"]\n")
		break
	endif
	sleep 50ms
endwhile
`

var connectTemplate *template.Template

func init() {
	connectTemplate = template.Must(template.New("connect").Parse(connectScript))
}

// NewChildVimServer creates Vim server process and connect to go-server by addr.
func NewChildVimServer(addr string) (*Process, error) {
	tmpfile, err := connectTmpFile(addr)
	if err != nil {
		return nil, err
	}

	cmd, err := vimServerCmd([]string{"-S", tmpfile.Name()})
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	process := &Process{cmd: cmd, script: tmpfile}

	return process, nil
}

func vimServerCmd(extraArgs []string) (*exec.Cmd, error) {

	path, err := exec.LookPath("vim")
	if err != nil {
		return nil, fmt.Errorf("vim not found: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := &exec.Cmd{
		Path:   path,
		Args:   append([]string{path}, extraArgs...),
		Stdin:  bytes.NewReader(nil), // Avoid "Input is not from a terminal"
		Stdout: &stdout,
		Stderr: &stderr,
	}
	return cmd, nil
}

// Connect connects server to Vim by servername (:h --servername)
func Connect(addr, vimServerName string, server *Server) (*Client, error) {
	if !remote.IsServed(vimServerName) {
		return nil, fmt.Errorf("server not found in vim --serverlist: %v", vimServerName)

	}

	tmpfile, err := connectTmpFile(addr)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(
		"vim",
		"--servername", vimServerName,
		"--remote-expr", fmt.Sprintf("execute(':source %v')", tmpfile.Name()),
	)

	savedHandler := server.Handler

	h := &getCliHandler{
		handler: savedHandler,
		chCli:   make(chan *Client, 1),
	}
	server.Handler = h
	defer func() { server.Handler = savedHandler }()

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	select {
	case cli := <-h.chCli:
		return cli, nil
	case <-time.After(15 * time.Second):
		return nil, ErrTimeOut
	}
}

func connectTmpFile(addr string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "vim-go-client")
	if err != nil {
		return nil, err
	}
	defer tmpfile.Close()

	connectTemplate.Execute(tmpfile, struct{ Addr string }{addr})
	return tmpfile, nil
}
