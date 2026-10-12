package cmd

import (
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable [file]",
	Short: "Enable environment variable(s) by setting their value to true",
	Long: `Set one or more environment variables to "true" in a .env file.
The .env file should be named .env or start with .env. (e.g. .env.local).
If [file] is omitted, it defaults to .env in the current directory.

If the variable does not exist, it will be created.
If the variable already exists with a value other than empty, null, true, or false, confirmation will be requested before making the change.`,
	Example: `  godot enable --app-debug
  godot enable .env --statamic-pro-enabled
  godot enable .env.local --statamic-git-enabled --statamic-stache-watcher`,
	Args: cobra.MaximumNArgs(1),
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBooleanCommand(cmd, args, "true")
	},
}
