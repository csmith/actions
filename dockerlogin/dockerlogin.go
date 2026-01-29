package dockerlogin

import (
	"fmt"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, registry, username, password, authfile string) error {
	args := []string{
		"login",
		registry,
		"-u", username,
		"--password-stdin",
		"--authfile", authfile,
	}

	cmd := exec.Command("buildah", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start buildah login: %w", err)
	}

	if _, err := stdin.Write([]byte(password)); err != nil {
		return fmt.Errorf("failed to write password to stdin: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("buildah login failed: %w", err)
	}

	return ctx.WriteOutput(map[string]string{"authfile": ctx.ResolvePath(authfile)})
}
