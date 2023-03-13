// +build windows

package localcommand

import (
	"os/exec"
	"syscall"
	"time"
	"io"
	"log"

	"github.com/creack/pty"
	// "github.com/photostorm/pty"
	"github.com/pkg/errors"
)

const (
	DefaultCloseSignal  = syscall.SIGINT
	DefaultCloseTimeout = 10 * time.Second
)

type LocalCommand struct {
	command string
	argv    []string

	closeSignal  syscall.Signal
	closeTimeout time.Duration

	cmd       *exec.Cmd

	stdin	io.WriteCloser
	stdout	io.ReadCloser
	stderr	io.ReadCloser

	// stdcomb bytes.Buffer

	pty	pty.Pty
	ispty 	bool
	ptyClosed chan struct{}
}

func New(command string, argv []string, options ...Option) (*LocalCommand, error) {
	cmd := exec.Command(command, argv...)
	pty, err := pty.Start(cmd)
	lcmd := &LocalCommand{
		command: command,
		argv:    argv,

		closeSignal:  DefaultCloseSignal,
		closeTimeout: DefaultCloseTimeout,

		cmd:       cmd,
		stdin:     nil,
		stdout:    nil,
		//stderr:    stderr,
		pty:       pty,
		ispty:	true,
		ptyClosed: nil,
	}
	if err != nil {
		log.Printf("Failed to start command `%s` with conpty, trying without (limited experience)", command)
		// todo close cmd?
		lcmd.ispty = false
	        lcmd.stdin, _ = cmd.StdinPipe()
		lcmd.stdout, _ = cmd.StdoutPipe()
		// stderr, _ := cmd.StderrPipe()
		/// var stdcomb bytes.Buffer
		err2 := lcmd.cmd.Start()
		if err2 != nil {
			return nil, errors.Wrapf(err, "failed to start command even without conpty `%s`", command)
		}
	}
	lcmd.ptyClosed = make(chan struct{})

	for _, option := range options {
		option(lcmd)
	}

	// When the process is closed by the user,
	// close pty so that Read() on the pty breaks with an EOF.
	go func() {
		defer func() {
			if lcmd.ispty {
				lcmd.pty.Close()
			}
			close(lcmd.ptyClosed)
		}()
		lcmd.cmd.Wait()
	}()

	return lcmd, nil
}

func (lcmd *LocalCommand) Read(p []byte) (n int, err error) {
	if (lcmd.ispty) {
		return lcmd.pty.Read(p)
	}
	return lcmd.stdout.Read (p)
}

func (lcmd *LocalCommand) Write(p []byte) (n int, err error) {
	if (lcmd.ispty) {
		return lcmd.pty.Write(p)
	}
	return lcmd.stdin.Write(p)
}

func (lcmd *LocalCommand) Close() error {
	if lcmd.cmd != nil && lcmd.cmd.Process != nil {
		lcmd.cmd.Process.Signal(lcmd.closeSignal)
	}
	for {
		select {
		case <-lcmd.ptyClosed:
			return nil
		case <-lcmd.closeTimeoutC():
			lcmd.cmd.Process.Signal(syscall.SIGKILL)
		}
	}
}

func (lcmd *LocalCommand) WindowTitleVariables() map[string]interface{} {
	return map[string]interface{}{
		"command": lcmd.command,
		"argv":    lcmd.argv,
		"pid":     lcmd.cmd.Process.Pid,
	}
}

func (lcmd *LocalCommand) ResizeTerminal(width int, height int) error {
	// only conpty support resizing of terminal
	if lcmd.ispty {
		winsize:=&pty.Winsize{
			Rows: uint16(height),
			Cols: uint16(width),
		}
		errno:=pty.Setsize(lcmd.pty, winsize)
		return errno
	}
	return nil
}

func (lcmd *LocalCommand) closeTimeoutC() <-chan time.Time {
	if lcmd.closeTimeout >= 0 {
		return time.After(lcmd.closeTimeout)
	}

	return make(chan time.Time)
}
