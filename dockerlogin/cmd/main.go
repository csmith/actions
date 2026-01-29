package main

import (
	"flag"
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/dockerlogin"
)

var (
	registry = flag.String("registry", "", "Registry URL")
	username = flag.String("username", "", "Username for authentication")
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

	password, hasPassword := os.LookupEnv("PASSWORD")
	if !hasPassword {
		fmt.Fprintf(os.Stderr, "Error: PASSWORD environment variable not set\n")
		os.Exit(1)
	}

	if err := dockerlogin.Run(ctx, *registry, *username, password, *authfile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
