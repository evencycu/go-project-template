package main

import (
	"os"

	"gitlab.com/cake/go-project-template/command"
)

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
