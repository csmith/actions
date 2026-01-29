package dockerpush

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"chameth.com/actions/common"
)

const localImageTag = "dockerpush-image:latest"

func Run(ctx *common.Context, archive, name, tags, authfile string) error {
	if tags == "" {
		return fmt.Errorf("tags cannot be empty")
	}

	tagList := strings.Split(tags, ",")
	for i, tag := range tagList {
		tagList[i] = strings.TrimSpace(tag)
	}

	for _, tag := range tagList {
		if tag == "" {
			return fmt.Errorf("tags cannot contain empty values")
		}
	}

	slog.Info("Pushing Docker image",
		"archive", archive,
		"image_name", name,
		"tags", strings.Join(tagList, ","),
		"authfile", authfile != "",
	)

	resolvedArchive := ctx.ResolvePath(archive)

	slog.Debug("Loading image from archive", "archive", resolvedArchive)
	loadArgs := []string{"pull", fmt.Sprintf("oci-archive:%s", resolvedArchive), localImageTag}
	loadCmd := exec.Command("buildah", loadArgs...)
	loadCmd.Stdout = os.Stdout
	loadCmd.Stderr = os.Stderr
	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("buildah pull (from archive) failed: %w", err)
	}

	for _, tag := range tagList {
		target := fmt.Sprintf("%s:%s", name, tag)
		slog.Debug("Pushing tag", "target", target)

		args := []string{
			"push",
		}

		if authfile != "" {
			args = append(args, "--authfile", ctx.ResolvePath(authfile))
		}

		args = append(args, localImageTag, target)

		cmd := exec.Command("buildah", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("buildah push failed for tag %s: %w", tag, err)
		}

		slog.Info("Tag pushed successfully", "target", target)
	}

	slog.Info("All tags pushed successfully")
	return nil
}
