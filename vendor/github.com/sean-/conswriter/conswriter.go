package conswriter

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/sean-/pager"
)

type ConsoleWriter interface {
	io.Writer
	io.Closer
	Wait() error
	getPager() *pager.Pager
}

var mtx sync.Mutex
var globalTerm ConsoleWriter
var origStdout *os.File
var origStderr *os.File

type terminal struct {
	stdout io.Writer
	stderr io.Writer
}

func GetTerminal() ConsoleWriter {
	mtx.Lock()
	defer mtx.Unlock()

	return globalTerm
}

func UsePager(usePager bool) error {
	mtx.Lock()
	defer mtx.Unlock()

	gp := globalTerm.getPager()

	switch {
	case usePager && gp != nil:
		// nop
		return nil
	case usePager && gp == nil:
		// If we're using a pager, copy stdout to stderr
		os.Stderr = os.Stdout

		p, err := pager.New()
		if err != nil {
			return err
		}
		t := &terminal{
			stdout: p,
			stderr: p,
		}
		globalTerm = t
		return nil
	case !usePager:
		t := &terminal{
			stdout: origStdout,
			stderr: origStderr,
		}
		globalTerm = t
		os.Stdout = origStdout
		os.Stderr = origStderr
		return nil
	default:
		panic("x")
	}
	// if stdout is already a pager then nothing
	// if stdout is terminal then insert pager
	// if stdout is pager then replace with terminal
	// if stdout is a terminal then nothing
}

func (t *terminal) Close() error {
	if p := t.getPager(); p != nil {
		return p.Wait()
	}

	return nil
}

func (t *terminal) Wait() error {
	if p := t.getPager(); p != nil {
		return p.Wait()
	}

	return nil
}

func (t *terminal) getPager() *pager.Pager {
	if p, ok := t.stdout.(*pager.Pager); ok {
		return p
	}

	return nil
}
func (t *terminal) Write(data []byte) (n int, err error) {
	return t.stdout.Write(data)
}

func init() {
	mtx.Lock()
	defer mtx.Unlock()

	globalTerm = &terminal{
		stdout: zerolog.SyncWriter(os.Stdout),
		stderr: zerolog.SyncWriter(os.Stderr),
	}

	// keep a copy of Stdout/Stderr
	origStdout = os.Stdout
	origStderr = os.Stderr
}
