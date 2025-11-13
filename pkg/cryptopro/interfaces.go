package cryptopro

import "context"

type executeCommander interface {
	CommandRun(ctx context.Context, cmd string) error
}
