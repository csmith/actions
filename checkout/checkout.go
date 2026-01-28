package checkout

import (
	"fmt"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context) error {
	if err := os.MkdirAll(ctx.Workspace, 0755); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	if err := os.Chdir(ctx.Workspace); err != nil {
		return fmt.Errorf("failed to change to workspace: %w", err)
	}

	cmd := exec.Command("git", "-c", fmt.Sprintf("http.extraHeader=Authorization: basic %s", ctx.BasicAuth()), "clone", ctx.RepoUrl(), ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, output)
	}

	cmd = exec.Command("git", "checkout", ctx.SHA)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, output)
	}

	return nil
}
