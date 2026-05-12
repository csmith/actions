package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/wowaddon"
)

var (
	source      = flag.String("source", "src", "Source directory containing the addon")
	destination = flag.String("destination", ".", "Destination directory for the zip file")
	debug       = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	common.ConfigureLogging(*debug)

	if err := wowaddon.Run(ctx, *source, *destination); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
