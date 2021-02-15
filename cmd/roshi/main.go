package main

import (
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:     "roshi",
	Version: "v0.0.0",
}

func main() {
	cmd.Execute()
}
