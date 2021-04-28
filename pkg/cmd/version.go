package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildVersion string
	BuildCommit  string
	BuildDate    string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays version for the executable",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ocm-load-test: %s-%s Build date:%s\n", BuildVersion, BuildCommit, BuildDate)
	},
}

func NewVersionCommand() *cobra.Command {
	return versionCmd
}
