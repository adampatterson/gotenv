package cmd

import (
	"fmt"
	"strings"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm [file]",
	Aliases: []string{"delete"},
	Short:   "Delete variables entirely from a .env file",
	Long: `Remove one or more environment variables specified via flags from a .env file, deleting the variable line entirely.
If [file] is omitted, it defaults to .env in the current directory.
To keep the variable but blank its value instead, use "unset".`,
	Example: `  godot rm .env --app-key
  godot delete --statamic-license-key --statamic-pro-enabled
  godot rm .env.local --statamic-git-enabled`,
	Args: cobra.MaximumNArgs(1),
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := ".env"
		if len(args) > 0 {
			filePath = args[0]
		}

		envFile, err := ResolveEnvFilePath(filePath)
		if err != nil {
			return err
		}

		_, envText, err := gotenv.ReadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("Unable to read %s file: %w", envFile, err)
		}

		// Extract flags to delete
		envFlags, err := GetUnknownFlags(cmd)
		if err != nil {
			return fmt.Errorf("Error reading command flags: %w", err)
		}

		if len(envFlags) < 1 {
			_ = cmd.Help()
			return nil
		}

		var keysToRemove []string
		for flag := range envFlags {
			if !strings.HasPrefix(flag, "--") {
				return fmt.Errorf("The flag must be in long format, but received %s", flag)
			}
			envName, err := FlagToEnvKey(flag)
			if err != nil {
				return err
			}
			keysToRemove = append(keysToRemove, envName)
		}

		removedCount, err := gotenv.RemoveEnvFileKeys(envFile, keysToRemove, envText)
		if err != nil {
			return fmt.Errorf("Error updating %s: %w", envFile, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Removed %d variable(s) (%s) from %s\n", removedCount, strings.Join(keysToRemove, ", "), envFile)
		return nil
	},
}
