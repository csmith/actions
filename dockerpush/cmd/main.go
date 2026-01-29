package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/dockerpush"
)

var (
	archive  = flag.String("archive", "image.tar", "Path to the image tar file to push")
	name     = flag.String("name", "", "Base image name")
	tags     = flag.String("tags", "", "Comma-separated list of tags to push")
	authfile = flag.String("authfile", ".registry-auth.json", "Path to authentication file")
	debug    = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	common.ConfigureLogging(*debug)

	if err := dockerpush.Run(ctx, *archive, *name, *tags, *authfile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
