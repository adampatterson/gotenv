package cmd

import (
	"github.com/spf13/cobra"
)

// dotEnvCmd provides the `dotenv` sub-namespace matching DDEV syntax (e.g. `godot dotenv get ...`)
var dotEnvCmd = &cobra.Command{
	Use:   "dotenv [command]",
	Short: "Commands for managing the contents of .env files",
	Long:  `Subcommands for inspecting, getting, and setting values in .env files (compatible with DDEV dotenv syntax).`,
	Example: `  godot dotenv get --app-key
  godot dotenv set --app-name="Statamic Peak"
  godot dotenv unset --app-key
  godot dotenv list`,
	Args: cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		envFlags, _ := GetUnknownFlags(cmd)
		if len(envFlags) == 0 && len(args) == 0 && commentFlag == "" {
			_ = cmd.Usage()
			return nil
		}
		return runSetCommand(cmd, args)
	},
}

func init() {
	dotEnvCmd.Flags().StringVarP(&commentFlag, "comment-text", "c", "", "Comment to add above the environment variable")
	dotEnvCmd.Flags().StringVar(&commentFlag, "with-comment", "", "Comment to add above the environment variable")

	dotEnvCmd.AddCommand(getCmd)
	dotEnvCmd.AddCommand(setCmd)
	dotEnvCmd.AddCommand(unsetCmd)
	dotEnvCmd.AddCommand(rmCmd)
	dotEnvCmd.AddCommand(listCmd)
	dotEnvCmd.AddCommand(enableCmd)
	dotEnvCmd.AddCommand(disableCmd)
	dotEnvCmd.AddCommand(skillCmd)
}
