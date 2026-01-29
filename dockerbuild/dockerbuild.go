package dockerbuild

import (
	"fmt"
	"os"
	"os/exec"

	"chameth.com/actions/common"
)

const imageTag = "dockerbuild-image:latest"

func Run(ctx *common.Context, dockerfile, context, target, authfile string) error {
	sourceLabel := fmt.Sprintf("%s/%s", ctx.ServerURL, ctx.Repository)

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

	args = append(args, ctx.ResolvePath(context))

	cmd := exec.Command("buildah", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildah build failed: %w", err)
	}

	targetPath := ctx.ResolvePath(target)

	pushArgs := []string{
		"push",
		"--format", "oci",
		imageTag,
		fmt.Sprintf("oci-archive:%s", targetPath),
	}

	if authfile != "" {
		pushArgs = append(pushArgs, "--authfile", authfile)
	}

	cmd = exec.Command("buildah", pushArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildah push failed: %w", err)
	}

	return ctx.WriteOutput(map[string]string{"image": target})
}
