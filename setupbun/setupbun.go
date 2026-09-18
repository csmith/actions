package setupbun

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"chameth.com/actions/common"
)

const bundledBunPath = "/usr/local/bin/bun"

func Run(ctx *common.Context, targetDir string) error {
	return install(ctx, targetDir, bundledBunPath)
}

func install(ctx *common.Context, targetDir string, source string) error {
	resolvedTarget := ctx.ResolvePath(targetDir)

	slog.Info("Setting up Bun", "target", resolvedTarget)

	binPath := resolvedTarget + "/bin"
	if err := os.MkdirAll(binPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	slog.Info("Copying Bun binary", "src", source, "dst", binPath+"/bun")
	if err := copyExecutable(source, binPath+"/bun"); err != nil {
		return fmt.Errorf("failed to copy Bun binary: %w", err)
	}

	slog.Info("Creating bunx alias", "path", binPath+"/bunx")
	if err := createBunxAlias(binPath); err != nil {
		return fmt.Errorf("failed to create bunx alias: %w", err)
	}

	slog.Info("Adding Bun to PATH", "path", binPath)
	if err := ctx.AddToPath(binPath); err != nil {
		return fmt.Errorf("failed to add to path: %w", err)
	}

	slog.Info("Bun setup complete")
	return nil
}

func copyExecutable(source string, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return os.Chmod(destination, 0755)
}

func createBunxAlias(binPath string) error {
	aliasPath := binPath + "/bunx"
	if err := os.Remove(aliasPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Symlink("bun", aliasPath)
}
