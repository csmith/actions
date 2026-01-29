package imagetags

import (
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"strings"

	"chameth.com/actions/common"
	"github.com/csmith/gitrefs"
	"github.com/hashicorp/go-version"
)

func Run(ctx *common.Context) error {
	slog.Info("Generating image tags", "ref", ctx.Ref)
	tags, err := tags(ctx)
	if err != nil {
		return err
	}

	if tags == nil {
		slog.Info("No tags generated")
		return nil
	}

	slog.Info("Generated tags", "tags", strings.Join(tags, ","))
	return ctx.WriteOutput(map[string]string{"tags": strings.Join(tags, ",")})
}

func tags(ctx *common.Context) ([]string, error) {
	if ctx.Ref == "refs/heads/main" || ctx.Ref == "refs/heads/master" {
		slog.Debug("Using dev tag for main/master branch")
		return []string{"dev"}, nil
	}

	if after, ok := strings.CutPrefix(ctx.Ref, "refs/tags/"); ok {
		tag := strings.TrimPrefix(after, "v")
		slog.Debug("Processing version tag", "tag", tag)
		targetVersion, err := version.NewVersion(tag)

		tags, err := gitrefs.Fetch(ctx.RepoUrl(), gitrefs.WithAuth("x-access-token", ctx.Token), gitrefs.TagsOnly())
		if err != nil {
			return nil, fmt.Errorf("couldn't find tags for repository: %w", err)
		}

		slog.Debug("Fetched repository tags for version resolution", "count", len(tags))
		return resolve(targetVersion, parseVersions(maps.Keys(tags))), nil
	}

	return nil, nil
}

func parseVersions(input iter.Seq[string]) []*version.Version {
	var res []*version.Version
	for i := range input {
		v, err := version.NewVersion(strings.TrimPrefix(i, "v"))
		if err != nil {
			slog.Warn("Failed to parse repository tag", "tag", i, "error", err)
			continue
		}

		if v.Metadata() == "" && v.Prerelease() == "" {
			res = append(res, v)
		}
	}
	return res
}

func resolve(targetVersion *version.Version, availableVersions []*version.Version) []string {
	if targetVersion.Metadata() != "" || targetVersion.Prerelease() != "" {
		slog.Info("Tagged version is not ordinary release, just using tag directly", "version", targetVersion.Original(), "metadata", targetVersion.Metadata(), "prerelease", targetVersion.Prerelease())
		return []string{targetVersion.String()}
	}

	res := []string{targetVersion.String()}
	hasNewerMajor := false
	hasNewerMinor := false
	hasNewerPatch := false
	targetSegments := targetVersion.Segments()
	for _, v := range availableVersions {
		segments := v.Segments()

		if segments[0] > targetSegments[0] {
			hasNewerMajor = true
		}

		if segments[0] == targetSegments[0] && segments[1] > targetSegments[1] {
			hasNewerMinor = true
		}

		if segments[0] == targetSegments[0] && segments[1] == targetSegments[1] && segments[2] > targetSegments[2] {
			hasNewerPatch = true
		}
	}

	if !hasNewerPatch {
		res = append(res, fmt.Sprintf("%d.%d", targetSegments[0], targetSegments[1]))

		if !hasNewerMinor {
			res = append(res, fmt.Sprintf("%d", targetSegments[0]))

			if !hasNewerMajor {
				res = append(res, "latest")
			}
		}
	}

	return res
}
