package main

import (
	"context"
	"os"

	stellar "go.polydawn.net/stellar/cmd/stellar/app"
)

func main() {
	os.Exit(stellar.Main(context.Background(), os.Args, os.Stdin, os.Stdout, os.Stderr))
}
