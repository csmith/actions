package githubrelease

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"chameth.com/actions/common"
	"github.com/google/go-github/v92/github"
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

	client, err := github.NewClient(github.WithHTTPClient(http.DefaultClient), github.WithAuthToken(token))
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	if err := waitForTag(client, owner, name, tag, ctx.SHA, 5*time.Second, 30*time.Minute); err != nil {
		return err
	}

	rel, _, err := client.Repositories.CreateRelease(context.Background(), owner, name, github.CreateReleaseRequest{
		TagName:    tag,
		Name:       github.Ptr(strings.TrimPrefix(tag, "v")),
		Body:       github.Ptr(body),
		MakeLatest: github.Ptr("legacy"),
	})

	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	slog.Info("Created GitHub release", "url", rel.HTMLURL, "version", tag)

	if assets != "" {
		if err := uploadAssets(ctx, client, owner, name, rel.GetID(), assets); err != nil {
			return fmt.Errorf("failed to upload assets: %w", err)
		}
	}

	return nil
}

// waitForTag blocks until the tag exists on GitHub and points at sha. GitHub
// creates a missing tag on the default branch when a release is made, so the
// release must not be created until the push mirror has copied the tag over.
func waitForTag(client *github.Client, owner, repo, tag, sha string, delay, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var lastErr error
	for {
		wait := delay
		commit, err := tagCommit(ctx, client, owner, repo, tag)
		switch {
		case err != nil && ctx.Err() != nil:
			return timeoutError(tag, timeout, err)
		case err != nil:
			retryAfter, ok := transientRetryAfter(err)
			if !ok {
				return err
			}
			lastErr = err
			wait = max(wait, retryAfter)
			slog.Warn("Failed to check tag on GitHub", "tag", tag, "error", err, "retry_in", wait)
		case commit == sha:
			return nil
		case commit != "":
			return fmt.Errorf("tag %s on GitHub points at %s, but the workflow ran for %s", tag, commit, sha)
		default:
			lastErr = nil
			slog.Info("Waiting for tag to appear on GitHub", "tag", tag, "retry_in", wait)
		}

		select {
		case <-ctx.Done():
			return timeoutError(tag, timeout, lastErr)
		case <-time.After(wait):
		}
		delay *= 2
	}
}

func timeoutError(tag string, timeout time.Duration, lastErr error) error {
	if lastErr != nil {
		return fmt.Errorf("tag %s did not appear on GitHub within %s: %w", tag, timeout, lastErr)
	}
	return fmt.Errorf("tag %s did not appear on GitHub within %s", tag, timeout)
}

// transientRetryAfter reports whether err is worth retrying, and how long
// GitHub asked us to wait first, if it said.
func transientRetryAfter(err error) (time.Duration, bool) {
	if rateErr, ok := errors.AsType[*github.RateLimitError](err); ok {
		return time.Until(rateErr.Rate.Reset.Time), true
	}
	if abuseErr, ok := errors.AsType[*github.AbuseRateLimitError](err); ok {
		return abuseErr.GetRetryAfter(), true
	}
	if respErr, ok := errors.AsType[*github.ErrorResponse](err); ok {
		status := respErr.Response.StatusCode
		if status != http.StatusTooManyRequests && status < http.StatusInternalServerError {
			return 0, false
		}
		seconds, _ := strconv.Atoi(respErr.Response.Header.Get("Retry-After"))
		return time.Duration(seconds) * time.Second, true
	}
	if _, ok := errors.AsType[*url.Error](err); ok {
		return 0, true
	}
	return 0, false
}

// tagCommit returns the commit the tag points at, peeling annotated tags, or
// an empty string if the tag doesn't exist.
func tagCommit(ctx context.Context, client *github.Client, owner, repo, tag string) (string, error) {
	ref, _, err := client.Git.GetRef(ctx, owner, repo, "tags/"+tag)
	if err != nil {
		errResp, ok := errors.AsType[*github.ErrorResponse](err)
		if ok && errResp.Response.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to get tag %s: %w", tag, err)
	}

	obj := ref.GetObject()
	for obj.GetType() == "tag" {
		t, _, err := client.Git.GetTag(ctx, owner, repo, obj.GetSHA())
		if err != nil {
			return "", fmt.Errorf("failed to get tag object %s: %w", obj.GetSHA(), err)
		}
		obj = t.GetObject()
	}

	if obj.GetType() != "commit" {
		return "", fmt.Errorf("tag %s on GitHub points at a %s, not a commit", tag, obj.GetType())
	}
	return obj.GetSHA(), nil
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
