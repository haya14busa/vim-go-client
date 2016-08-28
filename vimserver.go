package vim

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
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

func Connect(addr, vimServerName string) error {
	tmpfile, err := connectTmpFile(addr)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"vim",
		"--servername", vimServerName,
		"--remote-expr", fmt.Sprintf("execute(':source %v')", tmpfile.Name()),
	)

	return cmd.Run()
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
