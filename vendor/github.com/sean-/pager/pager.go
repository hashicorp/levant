// The Pager package allows the program to easily pipe it's
// standard output through a Pager program
// (like how the man command does).
//
// Borrowed from: https://gist.github.com/dchapes/1d0c538ce07902b76c75 and
// reworked slightly.

package pager

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Pager struct {
	cmd  *exec.Cmd
	file io.WriteCloser
}

var pager Pager

// The environment variables to check for the name of (and arguments to)
// the Pager to run.
var PagerEnvVariables = []string{"PAGER"}

// The command names in $PATH to look for if none of the environment
// variables are set.
// Cannot include arguments.
var PagerCommands = []string{"less", "more"}

func pagerExecPath() (pagerPath string, args []string, err error) {
	for _, testVar := range PagerEnvVariables {
		pagerPath = os.Getenv(testVar)
		if pagerPath != "" {
			args = strings.Fields(pagerPath)
			if len(args) > 1 {
				return args[0], args[1:], nil
			}
		}
	}

	// This default only gets used if PagerCommands is empty.
	err = exec.ErrNotFound
	for _, testPath := range PagerCommands {
		pagerPath, err = exec.LookPath(testPath)
		if err == nil {
			switch {
			case path.Base(pagerPath) == "less":
				// TODO(seanc@): Make the buffer size conditional
				args := []string{"-X", "-F", "-R", "--buffers=65535"}
				return pagerPath, args, nil
			default:
				return pagerPath, nil, nil
			}
		}
	}
	return "", nil, err
}

// New returns a new io.WriteCloser connected to a Pager.
// The returned WriteCloser can be used as a replacement to os.Stdout,
// everything written to it is piped to a Pager.
// To determine what Pager to run, the environment variables listed
// in PagerEnvVariables are checked.
// If all are empty/unset then the commands listed in PagerCommands
// are looked for in $PATH.
func New() (*Pager, error) {
	p := &Pager{}
	if p.cmd != nil {
		return nil, errors.New("Pager: already exists")
	}
	pagerPath, args, err := pagerExecPath()
	if err != nil {
		return nil, err
	}

	// If the Pager is less(1), set some useful options
	switch {
	case path.Base(pagerPath) == "less":
		os.Setenv("LESSSECURE", "1")
	}

	p.cmd = exec.Command(pagerPath, args...)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr
	w, err := p.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	f, ok := w.(io.WriteCloser)
	if !ok {
		return nil, errors.New("Pager: exec.Command.StdinPipe did not return type io.WriteCloser")
	}
	p.file = f
	err = p.cmd.Start()
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Wait closes the pipe to the Pager setup with New() or Stdout() and waits
// for it to exit.
//
// This should normally be called before the program exists,
// typically via a defer call in main().
func (p *Pager) Wait() error {
	if p.cmd == nil {
		return nil
	}
	p.file.Close()
	return p.cmd.Wait()
}

func (p *Pager) Close() error {
	return nil
}

func (p *Pager) Write(data []byte) (n int, err error) {
	return p.file.Write(data)
}
