package dockerpush

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, archive, name, tags, authfile string) error {
	if tags == "" {
		return fmt.Errorf("tags cannot be empty")
	}

	tagList := strings.Split(tags, ",")
	for i, tag := range tagList {
		tagList[i] = strings.TrimSpace(tag)
	}

	slog.Info("Pushing image",
		"archive", archive,
		"image_name", name,
		"tags", strings.Join(tagList, ","),
		"authfile", authfile != "",
	)

	resolvedArchive := ctx.ResolvePath(archive)

	for _, tag := range tagList {
		target := fmt.Sprintf("%s:%s", name, tag)
		slog.Debug("Pushing tag", "target", target)

		args := []string{
			"copy",
		}

		if authfile != "" {
			args = append(args, "--authfile", ctx.ResolvePath(authfile))
		}

		args = append(args, fmt.Sprintf("oci-archive:%s", resolvedArchive), fmt.Sprintf("docker://%s", target))

		cmd := exec.Command("skopeo", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("skopeo copy failed for tag %s: %w", tag, err)
		}

		slog.Info("Tag pushed successfully", "target", target)
	}

	slog.Info("All tags pushed successfully")
	return nil
}
