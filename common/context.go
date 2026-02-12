package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Context struct {
	Workspace      string
	Token          string
	ServerURL      string
	Repository     string
	HeadRepository string
	Ref            string
	HeadRef        string
	SHA            string
	OutputFile     string
}

type pullRequestEvent struct {
	PullRequest struct {
		Head struct {
			Repo struct {
				FullName string `json:"full_name"`
			} `json:"repo"`
			Ref string `json:"ref"`
		} `json:"head"`
	} `json:"pull_request"`
}

func (c *Context) BasicAuth() string {
	return base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "x-access-token:%s", c.Token))
}

func (c *Context) RepoUrl() string {
	repo := c.Repository
	if c.HeadRepository != "" {
		repo = c.HeadRepository
	}
	return fmt.Sprintf("%s/%s.git", c.ServerURL, repo)
}

func (c *Context) ResolvePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return fmt.Sprintf("%s/%s", c.Workspace, path)
}

func (c *Context) Tag() string {
	if after, ok := strings.CutPrefix(c.Ref, "refs/tags/"); ok {
		return after
	}
	return ""
}

func (c *Context) WriteOutput(m map[string]string) error {
	f, err := os.OpenFile(c.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer f.Close()

	for k, v := range m {
		if _, err := fmt.Fprintf(f, "%s=%s\n", k, v); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

func ContextFromEnv() (*Context, error) {
	if _, hasForgejoJob := os.LookupEnv("FORGEJO_JOB"); hasForgejoJob {
		return contextFromEnv("FORGEJO")
	}
	if _, hasGitHubJob := os.LookupEnv("GITHUB_JOB"); hasGitHubJob {
		return contextFromEnv("GITHUB")
	}
	return nil, fmt.Errorf("unable to determine forge: neither FORGEJO_JOB nor GITHUB_JOB set")
}

func contextFromEnv(prefix string) (*Context, error) {
	ctx := &Context{
		Workspace:  lookupEnv(prefix, "WORKSPACE"),
		Token:      lookupEnv(prefix, "TOKEN"),
		ServerURL:  lookupEnv(prefix, "SERVER_URL"),
		Repository: lookupEnv(prefix, "REPOSITORY"),
		Ref:        lookupEnv(prefix, "REF"),
		SHA:        lookupEnv(prefix, "SHA"),
		OutputFile: lookupEnv(prefix, "OUTPUT"),
	}

	eventName := lookupEnv(prefix, "EVENT_NAME")
	if eventName == "pull_request" {
		eventPath := lookupEnv(prefix, "EVENT_PATH")
		if eventPath != "" {
			headRepo, headRef, err := parsePullRequestEvent(eventPath)
			if err != nil {
				slog.Warn("Failed to parse pull request event, using base repository", "error", err)
			} else {
				ctx.HeadRepository = headRepo
				ctx.HeadRef = headRef
			}
		}
	}

	return ctx, nil
}

func lookupEnv(prefix, key string) string {
	val, _ := os.LookupEnv(fmt.Sprintf("%s_%s", prefix, key))
	return val
}

func parsePullRequestEvent(path string) (headRepo, headRef string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read event file: %w", err)
	}

	var event pullRequestEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return "", "", fmt.Errorf("failed to parse event JSON: %w", err)
	}

	return event.PullRequest.Head.Repo.FullName, event.PullRequest.Head.Ref, nil
}
