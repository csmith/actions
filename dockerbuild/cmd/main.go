package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/dockerbuild"
)

var (
	dockerfile = flag.String("dockerfile", "", "Path to Dockerfile")
	context    = flag.String("context", ".", "Build context path")
	target     = flag.String("target", "image.tar", "Output tar file for the image")
	authfile   = flag.String("authfile", ".registry-auth.json", "Path to authfile")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	if err := dockerbuild.Run(ctx, *dockerfile, *context, *target, *authfile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
