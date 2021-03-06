package vim

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/kr/pty"
)

// Process represents Vim server process.
type Process struct {
	cmd *exec.Cmd

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
func NewChildVimServer(addr string, args []string) (*Process, error) {
	tmpfile, err := connectTmpFile(addr)
	if err != nil {
		return nil, err
	}

	cmd, err := vimServerCmd(append([]string{"-S", tmpfile.Name()}, args...))
	if err != nil {
		return nil, err
	}

	// Emulate terminal to avoid "Input is not from a terminal"
	if _, err := pty.Start(cmd); err != nil {
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

	cmd := &exec.Cmd{
		Path: path,
		Args: append([]string{path}, extraArgs...),
	}
	return cmd, nil
}

func connectTmpFile(addr string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "vim-go-client")
	if err != nil {
		return nil, err
	}
	defer tmpfile.Close()

	if err := connectTemplate.Execute(tmpfile, struct{ Addr string }{addr}); err != nil {
		return nil, err
	}
	return tmpfile, nil
}
