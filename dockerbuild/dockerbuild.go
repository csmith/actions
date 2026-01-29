package dockerbuild

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, dockerfile, context string) error {
	sourceLabel := fmt.Sprintf("%s/%s", ctx.ServerURL, ctx.Repository)

	args := []string{
		"bud",
		"--timestamp=0",
		"--identity-label=false",
		"--label", fmt.Sprintf("org.opencontainers.image.source=%s", sourceLabel),
		"--iidfile", "/tmp/build-iidfile",
	}

	if dockerfile != "" {
		args = append(args, "-f", dockerfile)
	}

	args = append(args, context)

	cmd := exec.Command("buildah", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("buildah build failed: %w", err)
	}

	imageID, err := os.ReadFile("/tmp/build-iidfile")
	if err != nil {
		return fmt.Errorf("failed to read image ID file: %w", err)
	}

	imageIDStr := strings.TrimSpace(string(imageID))

	return ctx.WriteOutput(map[string]string{"imageid": imageIDStr})
}
