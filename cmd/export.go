package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/drossan/http2postman/internal/apidog"
	"github.com/drossan/http2postman/internal/converter"
	"github.com/drossan/http2postman/internal/fs"
	"github.com/drossan/http2postman/internal/model"
	"github.com/drossan/http2postman/internal/parser"
	"github.com/drossan/http2postman/internal/postman"
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

		// Collection name: flag or interactive prompt
		collectionName, _ := cmd.Flags().GetString("name")
		if collectionName == "" {
			var err error
			collectionName, err = promptCollectionName(cmd.InOrStdin())
			if err != nil {
				return err
			}
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

		// Resolve version bump: flag or interactive prompt
		var version string
		if output == "" {
			output = writer.OutputPath(".", collectionName)
			_, _, _, found := writer.ReadExistingVersion(fsys, output)

			bump, err := resolveBumpType(cmd, found)
			if err != nil {
				return err
			}

			_, version = writer.ResolveVersionedOutput(fsys, ".", collectionName, bump)
			force = true
		}

		collection := converter.HTTPFilesToCollection(files, collectionName, version, env)

		postmanWriter := writer.NewPostmanWriter(fsys)
		if err := postmanWriter.Write(collection, output, force); err != nil {
			return err
		}

		cmd.Printf("Collection exported to %s\n", output)

		// Push to Postman API if requested
		push, _ := cmd.Flags().GetBool("push")
		if push {
			apiKey := os.Getenv("POSTMAN_API_KEY")
			if apiKey == "" {
				return fmt.Errorf("POSTMAN_API_KEY environment variable is required for --push")
			}
			client := postman.NewClient(apiKey, &http.Client{})

			workspaceID, err := resolveWorkspace(cmd, client)
			if err != nil {
				return fmt.Errorf("selecting workspace: %w", err)
			}

			bumpStr, _ := cmd.Flags().GetString("bump")
			if bumpStr == "" {
				bumpStr = "minor"
			}

			result, pushErr := client.PushCollection(collection, workspaceID, bumpStr)
			if pushErr != nil {
				return fmt.Errorf("pushing to Postman: %w", pushErr)
			}
			if result.Created {
				cmd.Printf("Collection created in Postman v%s (UID: %s)\n", result.Version, result.UID)
			} else {
				cmd.Printf("Collection updated in Postman v%s (UID: %s)\n", result.Version, result.UID)
			}
		}

		// Push to Apidog if requested
		pushApidog, _ := cmd.Flags().GetBool("push-apidog")
		if pushApidog {
			apidogToken := os.Getenv("APIDOG_ACCESS_TOKEN")
			if apidogToken == "" {
				return fmt.Errorf("APIDOG_ACCESS_TOKEN environment variable is required for --push-apidog")
			}
			apidogClient := apidog.NewClient(apidogToken, &http.Client{})

			projectID, err := resolveApidogProject(cmd, apidogClient)
			if err != nil {
				return fmt.Errorf("selecting Apidog project: %w", err)
			}

			result, err := apidogClient.PushCollection(projectID, collection)
			if err != nil {
				return fmt.Errorf("pushing to Apidog: %w", err)
			}
			cmd.Printf("Synced to Apidog: %d endpoints created, %d updated\n",
				result.EndpointsCreated, result.EndpointsUpdated)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("output", "o", "", "Output file path (auto-generated if omitted)")
	exportCmd.Flags().StringP("name", "n", "", "Collection name (skips interactive prompt)")
	exportCmd.Flags().String("bump", "", "Version bump type: minor, patch, major (skips interactive prompt)")
	exportCmd.Flags().String("workspace", "", "Postman workspace ID (skips interactive prompt)")
	exportCmd.Flags().BoolP("force", "f", false, "Overwrite output file if exists")
	exportCmd.Flags().Bool("push", false, "Push to Postman (requires POSTMAN_API_KEY env var)")
	exportCmd.Flags().Bool("push-apidog", false, "Push to Apidog (requires APIDOG_ACCESS_TOKEN env var)")
	exportCmd.Flags().Int("apidog-project", 0, "Apidog project ID (skips interactive prompt)")
}

// resolveBumpType returns the bump type from --bump flag or interactive prompt.
func resolveBumpType(cmd *cobra.Command, hasExistingVersion bool) (writer.BumpType, error) {
	bumpFlag, _ := cmd.Flags().GetString("bump")
	if bumpFlag != "" {
		return parseBumpFlag(bumpFlag)
	}
	if !hasExistingVersion {
		return writer.BumpMinor, nil
	}
	return promptBumpType(cmd.InOrStdin())
}

func parseBumpFlag(value string) (writer.BumpType, error) {
	switch strings.ToLower(value) {
	case "minor":
		return writer.BumpMinor, nil
	case "patch":
		return writer.BumpPatch, nil
	case "major":
		return writer.BumpMajor, nil
	default:
		return 0, fmt.Errorf("invalid bump type: %q (expected minor, patch, or major)", value)
	}
}

// resolveWorkspace returns the workspace ID from --workspace flag or interactive prompt.
func resolveWorkspace(cmd *cobra.Command, client *postman.Client) (string, error) {
	wsFlag, _ := cmd.Flags().GetString("workspace")
	if wsFlag != "" {
		return wsFlag, nil
	}
	return selectWorkspace(client, cmd.InOrStdin())
}

func selectWorkspace(client *postman.Client, reader io.Reader) (string, error) {
	workspaces, err := client.ListWorkspaces()
	if err != nil {
		return "", err
	}
	if len(workspaces) == 0 {
		return "", nil
	}
	if len(workspaces) == 1 {
		return workspaces[0].ID, nil
	}

	fmt.Println("Select workspace:")
	for i, ws := range workspaces {
		fmt.Printf("  [%d] %s\n", i+1, ws.Name)
	}
	fmt.Print("Choice (default 1): ")

	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return workspaces[0].ID, nil
	}
	choice := strings.TrimSpace(scanner.Text())
	if choice == "" {
		return workspaces[0].ID, nil
	}

	idx := 0
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(workspaces) {
		return "", fmt.Errorf("invalid choice: %q", choice)
	}
	return workspaces[idx-1].ID, nil
}

func resolveApidogProject(cmd *cobra.Command, client *apidog.Client) (int, error) {
	projectFlag, _ := cmd.Flags().GetInt("apidog-project")
	if projectFlag != 0 {
		return projectFlag, nil
	}
	return selectApidogProject(client, cmd.InOrStdin())
}

func selectApidogProject(client *apidog.Client, reader io.Reader) (int, error) {
	projects, err := client.ListProjects()
	if err != nil {
		return 0, err
	}
	if len(projects) == 0 {
		return 0, fmt.Errorf("no Apidog projects found")
	}
	if len(projects) == 1 {
		return projects[0].ID, nil
	}

	fmt.Println("Select Apidog project:")
	for i, p := range projects {
		fmt.Printf("  [%d] %s\n", i+1, p.Name)
	}
	fmt.Print("Choice (default 1): ")

	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return projects[0].ID, nil
	}
	choice := strings.TrimSpace(scanner.Text())
	if choice == "" {
		return projects[0].ID, nil
	}

	idx := 0
	if _, err := fmt.Sscanf(choice, "%d", &idx); err != nil || idx < 1 || idx > len(projects) {
		return 0, fmt.Errorf("invalid choice: %q", choice)
	}
	return projects[idx-1].ID, nil
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
