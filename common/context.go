package common

import (
	"fmt"
	"os"
)

type Context struct {
	Workspace  string
	Token      string
	ServerURL  string
	Repository string
	Ref        string
	SHA        string
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
	}, nil
}

func lookupEnv(prefix, key string) string {
	val, _ := os.LookupEnv(fmt.Sprintf("%s_%s", prefix, key))
	return val
}
