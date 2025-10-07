package main

import (
	"os"

	"github.com/kyco/godevwatch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
