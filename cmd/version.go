package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of http2postman",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("http2postman %s (commit: %s)\n", appVersion, appCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
