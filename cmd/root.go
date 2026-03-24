package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "http2postman",
	Short: "A CLI tool to convert HTTP requests to Postman collections and vice versa",
	Long: `http2postman is a CLI tool that helps you convert directories of HTTP request files
to Postman collections and import Postman collections into directories of HTTP request files.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	appVersion = "dev"
	appCommit  = "none"
)

// SetVersionInfo sets version and commit info from build ldflags.
func SetVersionInfo(version, commit string) {
	appVersion = version
	appCommit = commit
}
