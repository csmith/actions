package dockerbuild

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

const manifestName = "localhost/dockerbuild"

func Run(ctx *common.Context, dockerfile, buildContext, target, authfile, platforms string) error {
	sourceLabel := fmt.Sprintf("%s/%s", ctx.ServerURL, ctx.Repository)
	contextPath := ctx.ResolvePath(buildContext)
	targetPath := ctx.ResolvePath(target)

	slog.Info(
		"Building Docker image",
		slog.String("context", contextPath),
		slog.String("dockerfile", dockerfile),
		slog.String("target", targetPath),
		slog.String("platforms", platforms),
		slog.String("source_label", sourceLabel),
		slog.Bool("authfile", authfile != ""),
	)

	if err := os.Chdir(contextPath); err != nil {
		return fmt.Errorf("failed to change to context directory %s: %w", contextPath, err)
	}

	args := []string{
		"build",
		"--timestamp=0",
		"--identity-label=false",
		"--label", "org.opencontainers.image.source=" + sourceLabel,
		"--platform", platforms,
		"--manifest", manifestName,
	}

	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}

	// Use "." as the build context since we're in the context directory
	args = append(args, ".")

	if err := buildah(args...); err != nil {
		return fmt.Errorf("buildah build failed: %w", err)
	}

	if err := buildah("manifest", "push", "--all", manifestName, "oci-archive:"+targetPath); err != nil {
		return fmt.Errorf("buildah manifest push failed: %w", err)
	}

	slog.Info("Docker image built", "image", target)
	return ctx.WriteOutput(map[string]string{"image": target})
}

func buildah(args ...string) error {
	slog.Debug("Executing buildah", "args", args)
	cmd := exec.CommandContext(context.Background(), "buildah", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
