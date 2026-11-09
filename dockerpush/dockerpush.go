package dockerpush

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"chameth.com/actions/common"
)

func Run(ctx *common.Context, archive, name, tags, authfile string) error {
	if tags == "" {
		return errors.New("tags cannot be empty")
	}

	tagList := strings.Split(tags, ",")
	for i, tag := range tagList {
		tagList[i] = strings.TrimSpace(tag)
	}

	slog.Info(
		"Pushing image",
		slog.String("archive", archive),
		slog.String("image_name", name),
		slog.String("tags", strings.Join(tagList, ",")),
		slog.Bool("authfile", authfile != ""),
	)

	resolvedArchive := ctx.ResolvePath(archive)

	for _, tag := range tagList {
		target := fmt.Sprintf("%s:%s", name, tag)
		slog.Debug("Pushing tag", "target", target)

		args := []string{
			"copy",
			"--all",
		}

		if authfile != "" {
			args = append(args, "--authfile", ctx.ResolvePath(authfile))
		}

		args = append(args, "oci-archive:"+resolvedArchive, "docker://"+target)

		cmd := exec.CommandContext(context.Background(), "skopeo", args...)
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
