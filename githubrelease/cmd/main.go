package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/githubrelease"
)

var (
	repo      = flag.String("repo", "", "Repository to create release in")
	changelog = flag.String("changelog", "src/CHANGELOG.md", "Path to the CHANGELOG to use for release notes")
	debug     = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	common.ConfigureLogging(*debug)

	token, ok := os.LookupEnv("TOKEN")
	if !ok || token == "" {
		fmt.Fprintf(os.Stderr, "Error: TOKEN environment variable not set\n")
		os.Exit(1)
	}

	if err := githubrelease.Run(ctx, *repo, *changelog, token); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
