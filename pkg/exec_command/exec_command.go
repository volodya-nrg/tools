package exec_command

import (
	"context"
	"fmt"
	"os/exec"
)

type ExecCommand struct{}

func (e ExecCommand) CommandRunAndOutput(ctx context.Context, cmd string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute cmd with combined output: %s", err)
	}
	return out, nil
}

func (e ExecCommand) CommandRun(ctx context.Context, cmd string) error {
	if err := exec.CommandContext(ctx, "sh", "-c", cmd).Run(); err != nil {
		return fmt.Errorf("failed to execute cmd: %s", err)
	}
	return nil
}

func NewExecCommand() *ExecCommand {
	return &ExecCommand{}
}
