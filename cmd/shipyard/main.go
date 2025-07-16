package main

import (
	"context"
	"os"

	"github.com/NatoNathan/shipyard/internal/cli"
	"github.com/charmbracelet/fang"
)

func main() {
	if err := fang.Execute(context.Background(), cli.RootCmd); err != nil {
		os.Exit(1)
	}
}
