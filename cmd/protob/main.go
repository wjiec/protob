package main

import (
	"context"
	"protob/internal/subcommand"
	"time"

	"github.com/spf13/cobra"
)

var (
	Version     = "0.0.0"
	GitRevision = "0000000"
	BuildTime   = time.Now().Format("2006/01/02")
)

func main() {
	root := cobra.Command{Use: "protob"}

	root.AddCommand(subcommand.Compile())
	root.AddCommand(subcommand.Install())
	root.AddCommand(subcommand.Version(Version, GitRevision, BuildTime))

	_ = root.ExecuteContext(context.TODO())
}
