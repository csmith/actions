package checkout

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, path string) error {
	targetDir := ctx.ResolvePath(path)
	slog.Info("Checking out repository", "repo", ctx.Repository, "sha", ctx.SHA, "target_dir", targetDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	if err := os.Chdir(targetDir); err != nil {
		return fmt.Errorf("failed to change to workspace: %w", err)
	}

	slog.Debug("Cloning repository", "url", ctx.RepoUrl())
	cmd := exec.Command("git", "-c", fmt.Sprintf("http.extraHeader=Authorization: basic %s", ctx.BasicAuth()), "clone", ctx.RepoUrl(), ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, output)
	}

	slog.Debug("Checking out SHA", "sha", ctx.SHA)
	cmd = exec.Command("git", "checkout", ctx.SHA)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, output)
	}

	slog.Info("Repository checked out successfully", "path", path)
	return ctx.WriteOutput(map[string]string{"path": path})
}
