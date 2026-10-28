package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("run tests use sh")
	}
}

func writeEnv(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestRunInjectsEnvFileIntoChild(t *testing.T) {
	skipOnWindows(t)
	envFile := writeEnv(t, t.TempDir(), ".env", "FOO=bar\nGREETING=\"hello world\"\n")

	out, err := executeCommand("run", "--env-file", envFile, "--", "sh", "-c", `printf '%s|%s' "$FOO" "$GREETING"`)
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "bar|hello world", out)
}

func TestRunDefaultsToDotEnvInCurrentDirectory(t *testing.T) {
	skipOnWindows(t)
	dir := t.TempDir()
	writeEnv(t, dir, ".env", "FOO=from-cwd\n")
	t.Chdir(dir)

	out, err := executeCommand("run", "--", "sh", "-c", `printf '%s' "$FOO"`)
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "from-cwd", out)
}

func TestRunFileValuesOverrideParentEnvironment(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("FOO", "parent")
	t.Setenv("UNTOUCHED", "kept")
	envFile := writeEnv(t, t.TempDir(), ".env", "FOO=file\n")

	out, err := executeCommand("run", "--env-file", envFile, "--", "sh", "-c", `printf '%s|%s' "$FOO" "$UNTOUCHED"`)
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "file|kept", out)
}

func TestRunLaterEnvFileWins(t *testing.T) {
	skipOnWindows(t)
	dir := t.TempDir()
	base := writeEnv(t, dir, ".env", "FOO=base\nONLY_BASE=1\n")
	local := writeEnv(t, dir, ".env.local", "FOO=local\n")

	out, err := executeCommand("run", "--env-file", base, "--env-file", local, "--", "sh", "-c", `printf '%s|%s' "$FOO" "$ONLY_BASE"`)
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "local|1", out)
}

func TestRunAcceptsCommandWithoutDashes(t *testing.T) {
	skipOnWindows(t)
	envFile := writeEnv(t, t.TempDir(), ".env", "FOO=bar\n")

	out, err := executeCommand("run", "--env-file", envFile, "sh", "-c", `printf '%s' "$FOO"`)
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "bar", out)
}

func TestRunPropagatesChildExitCode(t *testing.T) {
	skipOnWindows(t)
	envFile := writeEnv(t, t.TempDir(), ".env", "FOO=bar\n")

	_, err := executeCommand("run", "--env-file", envFile, "--", "sh", "-c", "exit 3")
	require.Error(t, err)

	var exitErr *ExitCodeError
	require.True(t, errors.As(err, &exitErr), "err=%v", err)
	assert.Equal(t, 3, exitErr.Code)
}

func TestRunFailsWhenEnvFileIsMissing(t *testing.T) {
	skipOnWindows(t)
	dir := t.TempDir()
	marker := filepath.Join(dir, "ran")

	_, err := executeCommand("run", "--env-file", filepath.Join(dir, ".env"), "--", "sh", "-c", "touch "+marker)
	require.Error(t, err)
	assert.NoFileExists(t, marker)
}

func TestRunRequiresCommand(t *testing.T) {
	envFile := writeEnv(t, t.TempDir(), ".env", "FOO=bar\n")

	_, err := executeCommand("run", "--env-file", envFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command")
}

func TestRunReportsUnstartableCommand(t *testing.T) {
	envFile := writeEnv(t, t.TempDir(), ".env", "FOO=bar\n")

	_, err := executeCommand("run", "--env-file", envFile, "--", "gotenv-no-such-binary")
	require.Error(t, err)

	var exitErr *ExitCodeError
	assert.False(t, errors.As(err, &exitErr), "a missing binary is not a child exit code")
	assert.Contains(t, err.Error(), "gotenv-no-such-binary")
}

func TestExtractCommentFromArgsLeavesCommandAfterDashesAlone(t *testing.T) {
	args := []string{"run", "--", "echo", "comment=hi", "-c", "x", "-out", "y"}

	cleanArgs, comment := ExtractCommentFromArgs(args)
	assert.Equal(t, args, cleanArgs)
	assert.Empty(t, comment)
}
