package setupgo

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, targetDir string, gopath string) error {
	resolvedTarget := ctx.ResolvePath(targetDir)
	resolvedGopath := ctx.ResolvePath(gopath)

	slog.Info("Setting up Go", "target", resolvedTarget, "gopath", resolvedGopath)

	if err := os.MkdirAll(resolvedTarget, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	slog.Info("Copying Go installation", "src", "/usr/local/go", "dst", resolvedTarget)
	cmd := exec.Command("cp", "-a", "/usr/local/go/.", resolvedTarget)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy Go installation: %w\n%s", err, output)
	}

	goBinPath := resolvedTarget + "/bin"
	slog.Info("Adding Go to PATH", "path", goBinPath)
	if err := ctx.AddToPath(goBinPath); err != nil {
		return fmt.Errorf("failed to add to path: %w", err)
	}

	slog.Info("Creating GOPATH directory", "path", resolvedGopath)
	if err := os.MkdirAll(resolvedGopath, 0755); err != nil {
		return fmt.Errorf("failed to create gopath directory: %w", err)
	}

	slog.Info("Setting GOPATH", "path", resolvedGopath)
	if err := ctx.SetEnv("GOPATH", resolvedGopath); err != nil {
		return fmt.Errorf("failed to set GOPATH: %w", err)
	}

	slog.Info("Go setup complete")
	return nil
}
