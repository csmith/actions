package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/curseforge"
)

var (
	projectID = flag.String("project-id", "", "CurseForge project ID")
	path      = flag.String("path", "", "Path to the zip file to upload (supports glob patterns)")
	changelog = flag.String("changelog", "src/CHANGELOG.md", "Path to a changelog file to include with the upload")
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

	apiToken, ok := os.LookupEnv("API_TOKEN")
	if !ok || apiToken == "" {
		fmt.Fprintf(os.Stderr, "Error: API_TOKEN environment variable not set\n")
		os.Exit(1)
	}

	if *projectID == "" {
		fmt.Fprintf(os.Stderr, "Error: -project-id is required\n")
		os.Exit(1)
	}

	if *path == "" {
		fmt.Fprintf(os.Stderr, "Error: -path is required\n")
		os.Exit(1)
	}

	if err := curseforge.Run(ctx, apiToken, *projectID, *path, *changelog); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
