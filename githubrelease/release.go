package githubrelease

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/actions/common"
	"github.com/google/go-github/v82/github"
	"github.com/parkr/changelog"
)

func Run(ctx *common.Context, repo, filename, token string) error {
	cl, err := changelog.NewChangelogFromFile(filename)
	if err != nil {
		return fmt.Errorf("failed to parse changelog: %w", err)
	}

	tag := ctx.Tag()
	if tag == "" {
		return fmt.Errorf("unable to determine tag for ref %s", ctx.Ref)
	}

	version := findVersion(cl.Versions, tag)
	if version == nil {
		slog.Warn("No changelog entry found for version", "tag", tag)
		version = &changelog.Version{
			Version: tag,
		}
	}

	owner, name, _ := strings.Cut(repo, "/")

	client := github.NewClient(http.DefaultClient).WithAuthToken(token)
	rel, _, err := client.Repositories.CreateRelease(context.Background(), owner, name, &github.RepositoryRelease{
		TagName:    github.Ptr(tag),
		Name:       &version.Version,
		Body:       github.Ptr(version.String()),
		MakeLatest: github.Ptr("legacy"),
	})

	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	slog.Info("Created GitHub release", "url", rel.HTMLURL, "version", version.Version)

	return nil
}

func findVersion(versions []*changelog.Version, target string) *changelog.Version {
	for i := range versions {
		if sameVersion(versions[i].Version, target) {
			return versions[i]
		}
	}
	return nil
}

func sameVersion(candidate, target string) bool {
	return strings.TrimPrefix(candidate, "v") == strings.TrimPrefix(target, "v")
}
