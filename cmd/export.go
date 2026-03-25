package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/drossan/http2postman/internal/converter"
	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
	"github.com/drossan/http2postman/internal/parser"
	"github.com/drossan/http2postman/internal/writer"
)

var exportCmd = &cobra.Command{
	Use:   "export [directory]",
	Short: "Export HTTP requests to a Postman collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		output, _ := cmd.Flags().GetString("output")
		force, _ := cmd.Flags().GetBool("force")

		collectionName, err := promptCollectionName(cmd.InOrStdin())
		if err != nil {
			return err
		}

		fsys := fs.NewOSFileSystem()

		httpParser := parser.NewHTTPFileParser(fsys)
		files, err := httpParser.ParseDirectory(dir)
		if err != nil {
			return fmt.Errorf("parsing directory %s: %w", dir, err)
		}

		var env *model.Environment
		envPath, err := parser.FindEnvFile(fsys, dir)
		if err == nil {
			parsedEnv, err := parser.ParseEnvironmentFromFile(fsys, envPath)
			if err != nil {
				return fmt.Errorf("parsing environment file: %w", err)
			}
			env = &parsedEnv
		}

		// Auto-generate versioned output path if not explicitly set
		var version string
		if output == "" {
			_, _, _, found := writer.FindLatestVersion(fsys, ".", collectionName)
			bump := writer.BumpMinor
			if found {
				var err error
				bump, err = promptBumpType(cmd.InOrStdin())
				if err != nil {
					return err
				}
			}
			output, version = writer.ResolveVersionedOutput(fsys, ".", collectionName, bump)
		}

		collection := converter.HTTPFilesToCollection(files, collectionName, version, env)

		postmanWriter := writer.NewPostmanWriter(fsys)
		if err := postmanWriter.Write(collection, output, force); err != nil {
			return err
		}

		cmd.Printf("Collection exported to %s\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("output", "o", "", "Output file path (auto-generated with versioning if omitted)")
	exportCmd.Flags().BoolP("force", "f", false, "Overwrite output file if exists")
}

func promptBumpType(reader io.Reader) (writer.BumpType, error) {
	fmt.Print("Version bump — [1] minor (default), [2] patch, [3] major: ")
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return writer.BumpMinor, nil
	}
	choice := strings.TrimSpace(scanner.Text())
	switch choice {
	case "", "1":
		return writer.BumpMinor, nil
	case "2":
		return writer.BumpPatch, nil
	case "3":
		return writer.BumpMajor, nil
	default:
		return 0, fmt.Errorf("invalid choice: %q (expected 1, 2, or 3)", choice)
	}
}

func promptCollectionName(reader io.Reader) (string, error) {
	fmt.Print("Enter the name for the Postman collection: ")
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		return "", model.ErrEmptyCollectionName
	}
	name := strings.TrimSpace(scanner.Text())
	if name == "" {
		return "", model.ErrEmptyCollectionName
	}
	return name, nil
}
