package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/checkout"
	"chameth.com/actions/common"
)

var (
	path  = flag.String("path", "src", "Path to checkout to (relative to workspace)")
	debug = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	common.ConfigureLogging(*debug)

	if err := checkout.Run(ctx, *path); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
