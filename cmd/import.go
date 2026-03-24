package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/drossan/http2postman/internal/converter"
	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/parser"
	"github.com/drossan/http2postman/internal/writer"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import a Postman collection to HTTP files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		output, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")

		fsys := fs.NewOSFileSystem()

		collection, err := parser.ParsePostmanCollectionFromFile(fsys, file)
		if err != nil {
			return fmt.Errorf("reading collection: %w", err)
		}

		httpFiles := converter.CollectionToHTTPFiles(collection)

		httpWriter := writer.NewHTTPFileWriter(fsys)
		if err := httpWriter.Write(httpFiles, output, force); err != nil {
			return err
		}

		cmd.Printf("HTTP files created in %s/\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringP("output", "o", "http-requests", "Output directory path")
	importCmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
}
