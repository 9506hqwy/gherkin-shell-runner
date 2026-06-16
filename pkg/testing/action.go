package testing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/aymanbagabas/go-pty"
	"github.com/cucumber/godog"
)

const attachmentMime = "text/plain"

func execCommand(ctx context.Context) (context.Context, error) {
	t := getTuiFeature(ctx)
	return runFeature(ctx, t)
}

func runFeature(
	ctx context.Context,
	t *tuiFeature,
) (context.Context, error) {
	ptmx, err := pty.New()
	if err != nil {
		return ctx, err
	}

	err = setPty(t, &ptmx)
	if err != nil {
		ptmx.Close()
		return ctx, err
	}

	cmd, cancel, cmdCtxErr := createCommand(ctx, t, &ptmx)
	defer cancel()

	output, err := runCommand(t, &ptmx, cmd)
	// ptmx was closed in runCommand.

	if cmdCtxErr() == context.DeadlineExceeded {
		// cmd.Wait() do not returns context.DeadlineExceeded
		// when the context deadline is exceeded at unix system,
		// because cmd.Cancel() kill the process ?
		// https://github.com/golang/go/issues/21880
		ctx = setError(ctx, t, cmdCtxErr())
		ctx = context.WithValue(ctx, timeoutKey{}, true)
		err = nil
	}

	if err == context.DeadlineExceeded {
		ctx = setError(ctx, t, err)
		ctx = context.WithValue(ctx, timeoutKey{}, true)
		err = nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		ctx = setFail(ctx, t, exitErr)
		err = nil
	}

	if err != nil {
		ctx = setError(ctx, t, err)
		return ctx, err
	}

	t.exitCode = cmd.ProcessState.ExitCode()

	ctx = setOutput(ctx, t, output)

	return ctx, err
}

func createCommand(
	ctx context.Context,
	t *tuiFeature,
	ptmx *pty.Pty,
) (*pty.Cmd, context.CancelFunc, func() error) {
	deadline, cancel := context.WithTimeout(
		ctx,
		time.Duration(t.timeout)*time.Millisecond,
	)

	cmd := (*ptmx).CommandContext(deadline, t.command, t.args...)

	cmd.Env = os.Environ()
	for key, value := range t.envs {
		env := fmt.Sprintf("%s=%s", key, value)
		cmd.Env = append(cmd.Env, env)
	}

	if t.workspace != EmptyString {
		cmd.Dir = t.workspace
	}

	return cmd, cancel, func() error {
		return deadline.Err()
	}
}

func runCommand(
	t *tuiFeature,
	ptmx *pty.Pty,
	cmd *pty.Cmd,
) (*bytes.Buffer, error) {
	var wg sync.WaitGroup

	err := cmd.Start()
	if err != nil {
		(*ptmx).Close()
		return nil, err
	}

	terminal := newTerminal()
	wg.Go(func() {
		_, _ = terminal.Copy(*ptmx)
	})

	if len(t.stdin) != ZERO {
		err = inputStdin(t, ptmx)
		if err != nil {
			(*ptmx).Close()
			return nil, err
		}
	}

	err = errors.Join(cmd.Wait(), waitBufferingCompleted(ptmx, &wg))
	// ptmx was closed in waitBufferingCompleted.

	if t.wait != ZERO {
		time.Sleep(time.Duration(t.wait) * time.Millisecond)
	}

	return terminal.Buffer(), err
}

func setOutput(
	ctx context.Context,
	t *tuiFeature,
	output *bytes.Buffer,
) context.Context {
	body := output.Bytes()

	t.output = body

	attachment := godog.Attachment{
		Body:      body,
		FileName:  "output",
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}

func setFail(
	ctx context.Context,
	_ *tuiFeature,
	err *exec.ExitError,
) context.Context {
	body := err.Stderr

	attachment := godog.Attachment{
		Body:      body,
		FileName:  "fail",
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}

func setError(
	ctx context.Context,
	_ *tuiFeature,
	err error,
) context.Context {
	body := err.Error()

	attachment := godog.Attachment{
		Body:      []byte(body),
		FileName:  "error",
		MediaType: attachmentMime,
	}

	return godog.Attach(ctx, attachment)
}
