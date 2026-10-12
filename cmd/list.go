package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

var (
	listJSONFlag   bool
	listExportFlag bool
)

var listCmd = &cobra.Command{
	Use:     "list [file]",
	Aliases: []string{"show", "dump"},
	Short:   "List all environment variables in a .env file",
	Long: `Display all variables defined in a .env file.
If [file] is omitted, it defaults to .env in the current directory.`,
	Example: `  godot list
  godot list .env.local
  godot list --json
  godot list --export`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := ".env"
		if len(args) > 0 {
			filePath = args[0]
		}

		envFile, err := ResolveEnvFilePath(filePath)
		if err != nil {
			return err
		}

		envMap, envText, err := gotenv.ReadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("Unable to read %s file: %w", envFile, err)
		}

		if listJSONFlag {
			data, err := json.MarshalIndent(envMap, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		if listExportFlag {
			keys := make([]string, 0, len(envMap))
			for k := range envMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Fprintf(cmd.OutOrStdout(), "export %s=%s\n", k, gotenv.EscapeEnvFileValue(envMap[k]))
			}
			return nil
		}

		tableOutput := RenderListTable(envMap, envText)
		if tableOutput != "" {
			fmt.Fprintln(cmd.OutOrStdout(), tableOutput)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVarP(&listJSONFlag, "json", "j", false, "Output list formatted as JSON")
	listCmd.Flags().BoolVarP(&listExportFlag, "export", "e", false, "Output with export prefix (e.g. export KEY=\"value\")")
}
