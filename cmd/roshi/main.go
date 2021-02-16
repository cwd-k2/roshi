package main

import (
	"github.com/cwd-k2/roshi/internal/subcmd/initcmd"
	"github.com/cwd-k2/roshi/internal/subcmd/pullcmd"
	"github.com/cwd-k2/roshi/internal/subcmd/pushcmd"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:     "roshi",
	Version: "v0.0.0",
}

func init() {
	cmd.AddCommand(
		initcmd.CMD,
		pullcmd.CMD,
		pushcmd.CMD,
	)
}

func main() {
	cmd.Execute()
}
