package main

import (
	"fmt"
	"os"

	"chameth.com/actions/common"
	"chameth.com/actions/imagetags"
)

func main() {
	ctx, err := common.ContextFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := imagetags.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
