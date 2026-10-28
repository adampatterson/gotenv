package cmd

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gotenv",
	Aliases: []string{"godot"},
	Short:   "gotenv is a CLI tool for managing .env files",
	Long:    `gotenv is a command line utility to inspect, get, set, push (sync), and delete variables in .env files, inspired by DDEV's dotenv command.`,
	Version: "1.0.1",
	Args:    cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		envFlags, _ := GetUnknownFlags(cmd)
		if len(envFlags) == 0 && len(args) == 0 && commentFlag == "" {
			_ = cmd.Help()
			return nil
		}
		return runSetCommand(cmd, args)
	},
}

// handleError prints errors with fang's default styling, except for a child exit code from `run`,
// which the child has already reported on its own.
func handleError(w io.Writer, styles fang.Styles, err error) {
	var exitErr *ExitCodeError
	if errors.As(err, &exitErr) {
		return
	}
	fang.DefaultErrorHandler(w, styles, err)
}

// Execute executes the root command.
func Execute() {
	cleanArgs, comment := ExtractCommentFromArgs(os.Args[1:])
	if comment != "" {
		commentFlag = comment
	}
	rootCmd.SetArgs(cleanArgs)
	err := fang.Execute(context.Background(), rootCmd, fang.WithVersion(rootCmd.Version), fang.WithErrorHandler(handleError))
	if err != nil {
		var exitErr *ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically answer yes to prompts (e.g. using example file or confirming changes)")

	rootCmd.Flags().StringVarP(&commentFlag, "comment-text", "c", "", "Comment to add above the environment variable")
	rootCmd.Flags().StringVar(&commentFlag, "with-comment", "", "Comment to add above the environment variable")
	rootCmd.Flags().StringSliceVar(&andFlags, "and", nil, "Additional target .env file(s) to update with the same values")

	// Add subcommands
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(unsetCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(skillCmd)
	rootCmd.AddCommand(dotEnvCmd)
}

// GetRootCmd returns the root command instance (useful for testing)
func GetRootCmd() *cobra.Command {
	return rootCmd
}
