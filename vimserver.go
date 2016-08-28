package vim

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
	"time"

	"local/haya14busa/go-vim-server/remote"
)

// Process represents Vim server process.
type Process struct {
	cmd  *exec.Cmd
	done chan error

	script *os.File
}

func (p *Process) Close() error {
	return os.Remove(p.script.Name())
}

const connectScript = `
call ch_logfile('/tmp/vimchannellog', 'w')
while 1
	let g:vim_server_handler = ch_open('{{ .Addr }}')
	if ch_status(g:vim_server_handler) is# 'open'
		echo 'open!'
		call ch_sendexpr(g:vim_server_handler, 'connect')
		break
	endif
	sleep 50ms
endwhile
`

var connectTemplate *template.Template

func init() {
	connectTemplate = template.Must(template.New("connect").Parse(connectScript))
}

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
	tmpfile, err := ioutil.TempFile("", "go-vim-server")
	if err != nil {
		return nil, err
	}
	defer tmpfile.Close()

	connectTemplate.Execute(tmpfile, struct{ Addr string }{addr})
	return tmpfile, nil
}
