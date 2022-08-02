package encoder

import (
	"context"
	"io"
	"os"
	"os/exec"
)

func PipeCommand(
	ctx context.Context, target io.Writer,
	cmd string, args ...string,
) (io.WriteCloser, error) {

	return PipeCmd(target, exec.CommandContext(ctx, cmd, args...))
}

func PipeCmd(target io.Writer, cmd *exec.Cmd) (io.WriteCloser, error) {
	cmd.Stdout = target
	cmd.Stderr = os.Stderr

	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	pc := &pipeCmd{
		wr:  input,
		cmd: cmd,
	}

	return pc, nil
}

// pipeCmd is an io.WriteCloser that wraps the stdin pipe for exec.Cmd,
// and handles starting the Cmd and waiting for it to finish.
type pipeCmd struct {
	wr  io.WriteCloser
	cmd *exec.Cmd
	err error
}

func (c *pipeCmd) Write(buf []byte) (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	if c.cmd.Process == nil {
		c.err = c.cmd.Start()
		if c.err != nil {
			return 0, c.err
		}
	}
	n, err := c.wr.Write(buf)
	if err != nil && c.err == nil {
		c.err = err
	}
	return n, err
}

func (c *pipeCmd) Close() error {
	defer func() {
		if c.err == nil {
			c.err = os.ErrClosed
		}
	}()
	if c.cmd.ProcessState != nil {
		if c.err == nil {
			c.err = os.ErrClosed
		}
		return c.err
	}
	if err := c.wr.Close(); err != nil && c.err == nil {
		c.err = err
	}
	if err := c.cmd.Wait(); err != nil && c.err == nil {
		c.err = err
	}
	return c.err
}
