package cmd

import (
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable [file]",
	Short: "Disable environment variable(s) by setting their value to false",
	Long: `Set one or more environment variables to "false" in a .env file.
The .env file should be named .env or start with .env. (e.g. .env.local).
If [file] is omitted, it defaults to .env in the current directory.

If the variable does not exist, it will be created.
If the variable already exists with a value other than empty, null, true, or false, confirmation will be requested before making the change.`,
	Example: `  godot disable --app-debug
  godot disable .env --statamic-stache-watcher
  godot disable .env.local --statamic-git-enabled --statamic-pro-enabled`,
	Args: cobra.MaximumNArgs(1),
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBooleanCommand(cmd, args, "false")
	},
}
