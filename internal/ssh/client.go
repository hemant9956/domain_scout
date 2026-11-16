package ssh

import (
	"context"
	"fmt"

	gossh "golang.org/x/crypto/ssh"
)

type Client struct{ client *gossh.Client }

func (c *Client) Close() error { return c.client.Close() }

func (c *Client) Run(ctx context.Context, command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()

	type result struct {
		output []byte
		err    error
	}
	ch := make(chan result, 1)
	go func() {
		out, runErr := session.CombinedOutput(command)
		ch <- result{output: out, err: runErr}
	}()
	select {
	case <-ctx.Done():
		_ = session.Signal(gossh.SIGKILL)
		return "", ctx.Err()
	case res := <-ch:
		if res.err != nil {
			return string(res.output), fmt.Errorf("remote command failed: %w: %s", res.err, string(res.output))
		}
		return string(res.output), nil
	}
}
