package common

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

type Context struct {
	Workspace  string
	Token      string
	ServerURL  string
	Repository string
	Ref        string
	SHA        string
	OutputFile string
}

func (c *Context) BasicAuth() string {
	return base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "x-access-token:%s", c.Token))
}

func (c *Context) RepoUrl() string {
	return fmt.Sprintf("%s/%s.git", c.ServerURL, c.Repository)
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
	return &Context{
		Workspace:  lookupEnv(prefix, "WORKSPACE"),
		Token:      lookupEnv(prefix, "TOKEN"),
		ServerURL:  lookupEnv(prefix, "SERVER_URL"),
		Repository: lookupEnv(prefix, "REPOSITORY"),
		Ref:        lookupEnv(prefix, "REF"),
		SHA:        lookupEnv(prefix, "SHA"),
		OutputFile: lookupEnv(prefix, "OUTPUT"),
	}, nil
}

func lookupEnv(prefix, key string) string {
	val, _ := os.LookupEnv(fmt.Sprintf("%s_%s", prefix, key))
	return val
}
