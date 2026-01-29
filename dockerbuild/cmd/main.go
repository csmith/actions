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
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	if err := dockerbuild.Run(ctx, *dockerfile, *context); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
