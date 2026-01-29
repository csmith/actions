package dockerbuild

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

const imageTag = "dockerbuild-image:latest"

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

	args := []string{
		"bud",
		"--timestamp=0",
		"--identity-label=false",
		"--label", fmt.Sprintf("org.opencontainers.image.source=%s", sourceLabel),
		"--tag", imageTag,
	}

	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}

	args = append(args, contextPath)

	slog.Debug("Executing buildah build", "args", args)
	cmd := exec.Command("buildah", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildah build failed: %w", err)
	}

	pushArgs := []string{
		"push",
		"--format", "oci",
		imageTag,
		fmt.Sprintf("oci-archive:%s", targetPath),
	}

	if authfile != "" {
		pushArgs = append(pushArgs, "--authfile", authfile)
	}

	slog.Debug("Executing buildah push to archive", "args", pushArgs)
	cmd = exec.Command("buildah", pushArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildah push failed: %w", err)
	}

	slog.Info("Docker image built and archived", "image", target)
	return ctx.WriteOutput(map[string]string{"image": target})
}
