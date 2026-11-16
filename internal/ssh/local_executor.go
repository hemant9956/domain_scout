package ssh

import (
	"context"
	"fmt"
	"os/exec"
)

// LocalExecutor runs the same read-only inspection commands against the local
// machine. It lets the agent reuse the SSH discovery pipeline without opening
// an SSH connection.
type LocalExecutor struct{}

func NewLocalExecutor() *LocalExecutor { return &LocalExecutor{} }

func (e *LocalExecutor) Run(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return string(output), ctx.Err()
		}
		return string(output), fmt.Errorf("command failed: %w: %s", err, string(output))
	}
	return string(output), nil
}
