package githubrelease

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"chameth.com/actions/common"
	"github.com/google/go-github/v86/github"
)

func Run(ctx *common.Context, repo, filename, token, assets string) error {
	tag := ctx.Tag()
	if tag == "" {
		return fmt.Errorf("unable to determine tag for ref %s", ctx.Ref)
	}

	body, err := ctx.Changelog(filename, tag)
	if err != nil {
		return err
	}
	if body == "" {
		slog.Warn("No changelog entry found for version", "tag", tag)
	}

	owner, name, _ := strings.Cut(repo, "/")

	client := github.NewClient(http.DefaultClient).WithAuthToken(token)
	rel, _, err := client.Repositories.CreateRelease(context.Background(), owner, name, &github.RepositoryRelease{
		TagName:    github.Ptr(tag),
		Name:       github.Ptr(strings.TrimPrefix(tag, "v")),
		Body:       github.Ptr(body),
		MakeLatest: github.Ptr("legacy"),
	})

	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	slog.Info("Created GitHub release", "url", *rel.HTMLURL, "version", tag)

	if assets != "" {
		if err := uploadAssets(ctx, client, owner, name, rel.GetID(), assets); err != nil {
			return fmt.Errorf("failed to upload assets: %w", err)
		}
	}

	return nil
}

func uploadAssets(ctx *common.Context, client *github.Client, owner, repo string, releaseID int64, assets string) error {
	patterns := strings.SplitSeq(assets, ",")
	for pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		resolved := ctx.ResolvePath(pattern)
		matches, err := filepath.Glob(resolved)
		if err != nil {
			return fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
		}
		if len(matches) == 0 {
			slog.Warn("No files matched glob pattern", "pattern", pattern)
			continue
		}

		for _, match := range matches {
			if err := uploadAsset(client, owner, repo, releaseID, match); err != nil {
				return err
			}
		}
	}
	return nil
}

func uploadAsset(client *github.Client, owner, repo string, releaseID int64, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open asset %q: %w", path, err)
	}
	defer f.Close()

	name := filepath.Base(path)
	slog.Info("Uploading release asset", "name", name, "path", path)

	_, _, err = client.Repositories.UploadReleaseAsset(
		context.Background(),
		owner,
		repo,
		releaseID,
		&github.UploadOptions{Name: name},
		f,
	)
	if err != nil {
		return fmt.Errorf("failed to upload asset %q: %w", name, err)
	}

	slog.Info("Uploaded release asset", "name", name)
	return nil
}
