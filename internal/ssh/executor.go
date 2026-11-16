package ssh

import "context"

type Executor interface {
	Run(ctx context.Context, command string) (string, error)
}
