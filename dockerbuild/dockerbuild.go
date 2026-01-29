package dockerbuild

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, dockerfile, context, target, authfile string) error {
	sourceLabel := fmt.Sprintf("%s/%s", ctx.ServerURL, ctx.Repository)
	contextPath := ctx.ResolvePath(context)
	targetPath := ctx.ResolvePath(target)

	slog.Info("Building Docker image",
		"context", contextPath,
		"dockerfile", dockerfile,
		"target", targetPath,
		"source_label", sourceLabel,
		"authfile", authfile != "")

	if err := os.Chdir(contextPath); err != nil {
		return fmt.Errorf("failed to change to context directory %s: %w", contextPath, err)
	}

	args := []string{
		"bud",
		"--timestamp=0",
		"--identity-label=false",
		"--label", fmt.Sprintf("org.opencontainers.image.source=%s", sourceLabel),
		"--tag", fmt.Sprintf("oci-archive:%s", targetPath),
	}

	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}

	// Use "." as the build context since we're in the context directory
	args = append(args, ".")

	slog.Debug("Executing buildah build", "args", args, "cwd", contextPath)
	cmd := exec.Command("buildah", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildah build failed: %w", err)
	}

	slog.Info("Docker image built", "image", target)
	return ctx.WriteOutput(map[string]string{"image": target})
}
