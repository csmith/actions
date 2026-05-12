package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/wowinterface"
)

var (
	addonID = flag.String("addon-id", "", "WowInterface addon ID")
	path    = flag.String("path", "", "Path to the zip file to upload (supports glob patterns)")
	debug   = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	flag.Parse()
	common.ConfigureLogging(*debug)

	apiKey, ok := os.LookupEnv("API_KEY")
	if !ok || apiKey == "" {
		fmt.Fprintf(os.Stderr, "Error: API_KEY environment variable not set\n")
		os.Exit(1)
	}

	if *addonID == "" {
		fmt.Fprintf(os.Stderr, "Error: -addon-id is required\n")
		os.Exit(1)
	}

	if *path == "" {
		fmt.Fprintf(os.Stderr, "Error: -path is required\n")
		os.Exit(1)
	}

	if err := wowinterface.Run(ctx, apiKey, *addonID, *path); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
