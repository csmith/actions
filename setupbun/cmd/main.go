package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/setupbun"
)

var (
	target = flag.String("target", "tools/bun", "Directory to install Bun into")
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

	if err := setupbun.Run(ctx, *target); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
