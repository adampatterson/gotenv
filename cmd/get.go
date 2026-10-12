package cmd

import (
	"fmt"
	"strings"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

var getWithCommentsFlag bool

var getCmd = &cobra.Command{
	Use:   "get [file]",
	Short: "Get the value of an environment variable from a .env file",
	Long: `Retrieve the value of an environment variable specified via a long flag from a .env file.
The .env file should be named .env or start with .env. (e.g. .env.local).
If [file] is omitted, it defaults to .env in the current directory.`,
	Example: `  godot get .env --app-key
  godot get --app-name
  godot get .env.local --statamic-license-key
  godot get --statamic-pro-enabled
  godot get .env --app-key --with-comments
  godot get .env --app-key -c`,
	Args: cobra.MaximumNArgs(1),
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			getWithCommentsFlag = false
		}()

		filePath := ".env"
		if len(args) > 0 {
			filePath = args[0]
		}

		envFile, err := ResolveEnvFilePath(filePath)
		if err != nil {
			return err
		}

		// Extract unknown flags
		envFlags, err := GetUnknownFlags(cmd)
		if err != nil {
			return fmt.Errorf("Error reading command flags: %w", err)
		}

		if len(envFlags) < 1 {
			_ = cmd.Help()
			return nil
		}

		if len(envFlags) != 1 {
			return fmt.Errorf("Only one environment variable can be retrieved at a time.")
		}

		var rawFlag string
		for f := range envFlags {
			rawFlag = f
		}

		if !strings.HasPrefix(rawFlag, "--") {
			return fmt.Errorf("The flag must be in long format, but received %s", rawFlag)
		}

		envName, err := FlagToEnvKey(rawFlag)
		if err != nil {
			return err
		}

		// Read the .env file
		envMap, envText, err := gotenv.ReadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("Unable to read %s file: %w", envFile, err)
		}

		val, exists := envMap[envName]
		if !exists {
			return fmt.Errorf("The environment variable '%s' not found in %s", envName, envFile)
		}

		if getWithCommentsFlag {
			comment := gotenv.GetEnvVarComment(envText, envName)
			tableOutput := RenderGetWithCommentsTable(envName, val, comment)
			fmt.Fprintln(cmd.OutOrStdout(), tableOutput)
			return nil
		}

		// Print raw unescaped value
		fmt.Fprintln(cmd.OutOrStdout(), gotenv.FormatEnvValue(val))
		return nil
	},
}

func init() {
	getCmd.Flags().BoolVarP(&getWithCommentsFlag, "with-comments", "c", false, "Include comment when retrieving variable")
	getCmd.Flags().BoolVar(&getWithCommentsFlag, "with-comment", false, "Include comment when retrieving variable")
	getCmd.Flags().BoolVar(&getWithCommentsFlag, "comments", false, "Include comment when retrieving variable")
}
