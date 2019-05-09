package main

import (
	"context"
	"os"

	reach "go.polydawn.net/reach/cmd/reach/app"
)

func main() {
	os.Exit(reach.Main(context.Background(), os.Args, os.Stdin, os.Stdout, os.Stderr))
}
