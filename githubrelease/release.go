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
		Body:       github.Ptr(createBody(version)),
		MakeLatest: github.Ptr("legacy"),
	})

	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	slog.Info("Created GitHub release", "url", *rel.HTMLURL, "version", version.Version)

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

func createBody(v *changelog.Version) string {
	var lines []string
	if len(v.History) > 0 {
		historyStrs := make([]string, len(v.History))
		for i, history := range v.History {
			historyStrs[i] = history.String()
		}
		lines = append(lines, strings.Join(historyStrs, "\n"))
	}
	if len(v.Subsections) > 0 {
		subsectionsStrs := make([]string, len(v.Subsections))
		for i, subsection := range v.Subsections {
			subsectionsStrs[i] = subsection.String()
		}
		lines = append(lines, strings.Join(subsectionsStrs, "\n\n"))
	}
	return strings.Join(lines, "\n\n")
}
