package main

import (
	"context"
	"log"
	"os"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "unknown"
)

func main() {
	cmd := RootCmd()

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
