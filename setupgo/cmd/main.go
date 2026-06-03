package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/setupgo"
)

var (
	target = flag.String("target", "tools/go", "Directory to install Go into")
	gopath = flag.String("gopath", "cache/go", "Directory to use as GOPATH")
	debug  = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()

	common.ConfigureLogging(*debug)

	if err := setupgo.Run(ctx, *target, *gopath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
