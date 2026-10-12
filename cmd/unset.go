package cmd

import (
	"fmt"
	"strings"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

var unsetCmd = &cobra.Command{
	Use:   "unset [file]",
	Short: "Blank the value of variables in a .env file",
	Long: `Clear the value of one or more environment variables specified via flags from a .env file, keeping the variable (and its comments) in place.
If [file] is omitted, it defaults to .env in the current directory.
To remove the variable entirely, use "rm" or "delete" instead.`,
	Example: `  godot unset .env --app-key
  godot unset --statamic-license-key --statamic-pro-enabled
  godot unset .env.local --statamic-git-enabled`,
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

		// Extract flags to unset
		envFlags, err := GetUnknownFlags(cmd)
		if err != nil {
			return fmt.Errorf("Error reading command flags: %w", err)
		}

		if len(envFlags) < 1 {
			_ = cmd.Help()
			return nil
		}

		var keysToUnset []string
		for flag := range envFlags {
			if !strings.HasPrefix(flag, "--") {
				return fmt.Errorf("The flag must be in long format, but received %s", flag)
			}
			envName, err := FlagToEnvKey(flag)
			if err != nil {
				return err
			}
			keysToUnset = append(keysToUnset, envName)
		}

		blankedCount, err := gotenv.BlankEnvFileKeys(envFile, keysToUnset, envText)
		if err != nil {
			return fmt.Errorf("Error updating %s: %w", envFile, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Unset %d variable(s) (%s) in %s\n", blankedCount, strings.Join(keysToUnset, ", "), envFile)
		return nil
	},
}
