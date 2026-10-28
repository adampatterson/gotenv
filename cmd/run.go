package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

var runEnvFiles []string

// ExitCodeError carries the exit code of a child process started by `run`,
// so Execute can exit with the same code without printing an error.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("command exited with code %d", e.Code)
}

var runCmd = &cobra.Command{
	Use:   "run [--env-file=file]... [--] command [args...]",
	Short: "Run a command with the variables from .env files in its environment",
	Long: `Load one or more .env files and run a command with those variables added to its environment.
Values from the files override variables that are already set. If --env-file is given more than once,
later files override earlier ones. If --env-file is omitted, it defaults to .env in the current directory.
The files are only read. The command inherits stdin, stdout, stderr, and signals, and gotenv exits with its exit code.`,
	Example: `  godot run -- ddev k6 run
  godot run --env-file=.env.local -- npm test
  godot run --env-file=.env --env-file=.env.local -- printenv APP_KEY`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		envFiles := runEnvFiles
		defer func() {
			runEnvFiles = nil
		}()

		if len(args) == 0 {
			return fmt.Errorf("A command to run is required, for example: gotenv run -- ddev k6 run")
		}
		if len(envFiles) == 0 {
			envFiles = []string{".env"}
		}

		fileEnv := make([]string, 0)
		for _, rawPath := range envFiles {
			envFile, err := ResolveEnvFilePath(rawPath)
			if err != nil {
				return err
			}
			envMap, _, err := gotenv.ReadEnvFile(envFile)
			if err != nil {
				return fmt.Errorf("Unable to read %s file: %w", envFile, err)
			}
			for k, v := range envMap {
				fileEnv = append(fileEnv, k+"="+v)
			}
		}

		child := exec.Command(args[0], args[1:]...)
		// exec keeps the last value when a key repeats, so file values win over inherited ones.
		child.Env = append(os.Environ(), fileEnv...)
		child.Stdin = cmd.InOrStdin()
		child.Stdout = cmd.OutOrStdout()
		child.Stderr = cmd.ErrOrStderr()

		if err := child.Start(); err != nil {
			return fmt.Errorf("Unable to run %s: %w", args[0], err)
		}

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signals)
		go func() {
			for sig := range signals {
				_ = child.Process.Signal(sig)
			}
		}()

		if err := child.Wait(); err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return &ExitCodeError{Code: exitCodeOf(exitErr)}
			}
			return fmt.Errorf("Unable to run %s: %w", args[0], err)
		}
		return nil
	},
}

// exitCodeOf returns the exit code of a finished child, using the shell convention
// of 128+signal when the child was killed by a signal.
func exitCodeOf(exitErr *exec.ExitError) int {
	if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}
	return exitErr.ExitCode()
}

func init() {
	// Stop flag parsing at the first positional argument so flags meant for the child pass through.
	runCmd.Flags().SetInterspersed(false)
	runCmd.Flags().StringArrayVar(&runEnvFiles, "env-file", nil, "Path to a .env file to load (repeatable, defaults to .env)")
}
