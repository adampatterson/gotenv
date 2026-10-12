package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommandWithInput(input string, args ...string) (string, error) {
	cmd := GetRootCmd()
	buf := new(bytes.Buffer)
	if input != "" {
		cmd.SetIn(strings.NewReader(input))
	} else {
		cmd.SetIn(strings.NewReader(""))
	}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	commentFlag = ""
	getWithCommentsFlag = false
	yesFlag = false
	cleanArgs, comment := ExtractCommentFromArgs(args)
	if comment != "" {
		commentFlag = comment
	}
	cmd.SetArgs(cleanArgs)

	// In test, also pass cleanArgs to GetUnknownFlags
	os.Args = append([]string{"gotenv"}, cleanArgs...)

	err := cmd.Execute()
	return buf.String(), err
}

func executeCommand(args ...string) (string, error) {
	return executeCommandWithInput("", args...)
}

func TestCLIGetAndSet(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// Set initial values
	out, err := executeCommand("set", envFile, "--app-name", "TestApp", "--db-port", "3306")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Updated")
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"TestApp"`)
	assert.Contains(t, out, "DB_PORT")
	assert.Contains(t, out, `"3306"`)

	// Verify file exists
	require.FileExists(t, envFile)

	// Get value
	out, err = executeCommand("get", envFile, "--app-name")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "TestApp\n", out)

	out, err = executeCommand("get", envFile, "--db-port")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "3306\n", out)

	// Set value with special characters
	out, err = executeCommand("set", envFile, "--secret-key", `pass"word$with$special\nnewline`)
	require.NoError(t, err, "out=%s", out)

	out, err = executeCommand("get", envFile, "--secret-key")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, `pass\"word$with$special\nnewline`+"\n", out)

	// List
	out, err = executeCommand("list", envFile)
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, "TestApp")
	assert.NotContains(t, out, `"TestApp"`)
	assert.Contains(t, out, "DB_PORT")
	assert.Contains(t, out, "3306")

	// List JSON
	out, err = executeCommand("list", envFile, "--json")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, `"APP_NAME": "TestApp"`)

	// Unset (should blank the value but keep the variable)
	out, err = executeCommand("unset", envFile, "--db-port")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Unset 1 variable(s)")

	// Get unset value should now be empty, not an error
	out, err = executeCommand("get", envFile, "--db-port")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "\n", out)

	// Rm (should remove the variable entirely)
	out, err = executeCommand("rm", envFile, "--db-port")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Removed 1 variable(s)")

	// Get removed value should fail
	_, err = executeCommand("get", envFile, "--db-port")
	require.Error(t, err)
}

func TestCLIValidationErrors(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")
	err := os.WriteFile(envFile, []byte("FOO=bar\n"), 0644)
	require.NoError(t, err)

	// Test invalid file prefix
	wrongFile := filepath.Join(tempDir, "invalid.txt")
	_, err = executeCommand("get", wrongFile, "--foo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "should have .env prefix")

	// Test invalid flag casing/characters
	_, err = executeCommand("set", envFile, "--INVALID_FLAG", "val")
	require.Error(t, err)

	// Test multiple flags on get
	_, err = executeCommand("get", envFile, "--foo", "--bar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Only one environment variable can be retrieved at a time")

	// Test missing variable on get
	_, err = executeCommand("get", envFile, "--non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in")
}

func TestCLIDotenvSubcommand(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	out, err := executeCommand("dotenv", "set", envFile, "--test-var", "hello")
	require.NoError(t, err, "out=%s", out)

	out, err = executeCommand("dotenv", "get", envFile, "--test-var")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "hello\n", out)
}

func TestCLIGenerateRandomHex(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	out, err := executeCommand("set", envFile, "--my-secret", "@generate", "--my-random", "@random")
	require.NoError(t, err, "out=%s", out)

	// Should have generated 32-character hex strings
	outSecret, err := executeCommand("get", envFile, "--my-secret")
	require.NoError(t, err, "out=%s", outSecret)

	outRandom, err := executeCommand("get", envFile, "--my-random")
	require.NoError(t, err, "out=%s", outRandom)

	// Remove trailing newline
	secretStr := outSecret[:len(outSecret)-1]
	randomStr := outRandom[:len(outRandom)-1]

	assert.Len(t, secretStr, 32)
	assert.Len(t, randomStr, 32)
	assert.NotEqual(t, "@generate", secretStr)
	assert.NotEqual(t, "@random", randomStr)
	assert.NotEqual(t, secretStr, randomStr) // Extremely unlikely to collide
}

func TestCLIGenerateLaravelAppKey(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	out, err := executeCommand("set", envFile, "--app-key", "@generate", "--custom-app-key", "@key", "--another-key", "@base64")
	require.NoError(t, err, "out=%s", out)

	outAppKey, err := executeCommand("get", envFile, "--app-key")
	require.NoError(t, err, "out=%s", outAppKey)

	outCustomKey, err := executeCommand("get", envFile, "--custom-app-key")
	require.NoError(t, err, "out=%s", outCustomKey)

	outAnotherKey, err := executeCommand("get", envFile, "--another-key")
	require.NoError(t, err, "out=%s", outAnotherKey)

	appKeyStr := strings.TrimSpace(outAppKey)
	customKeyStr := strings.TrimSpace(outCustomKey)
	anotherKeyStr := strings.TrimSpace(outAnotherKey)

	// Verify prefix and length (base64: + 44 chars = 51 chars)
	assert.True(t, strings.HasPrefix(appKeyStr, "base64:"))
	assert.Len(t, appKeyStr, 51)
	rawAppKey := strings.TrimPrefix(appKeyStr, "base64:")
	decodedAppKey, err := base64.StdEncoding.DecodeString(rawAppKey)
	require.NoError(t, err)
	assert.Len(t, decodedAppKey, 32)

	assert.True(t, strings.HasPrefix(customKeyStr, "base64:"))
	assert.Len(t, customKeyStr, 51)
	rawCustomKey := strings.TrimPrefix(customKeyStr, "base64:")
	decodedCustomKey, err := base64.StdEncoding.DecodeString(rawCustomKey)
	require.NoError(t, err)
	assert.Len(t, decodedCustomKey, 32)

	assert.True(t, strings.HasPrefix(anotherKeyStr, "base64:"))
	assert.Len(t, anotherKeyStr, 51)
	rawAnotherKey := strings.TrimPrefix(anotherKeyStr, "base64:")
	decodedAnotherKey, err := base64.StdEncoding.DecodeString(rawAnotherKey)
	require.NoError(t, err)
	assert.Len(t, decodedAnotherKey, 32)

	// Verify uniqueness
	assert.NotEqual(t, appKeyStr, customKeyStr)

	// Verify setting APP_KEY with @random also produces base64 key
	out2, err := executeCommand("set", envFile, "--app-key", "@random")
	require.NoError(t, err, "out=%s", out2)

	outAppKey2, err := executeCommand("get", envFile, "--app-key")
	require.NoError(t, err, "out=%s", outAppKey2)
	appKeyStr2 := strings.TrimSpace(outAppKey2)
	assert.True(t, strings.HasPrefix(appKeyStr2, "base64:"))
	assert.Len(t, appKeyStr2, 51)
	assert.NotEqual(t, appKeyStr, appKeyStr2)

	// Verify setting explicit literal value is preserved
	out3, err := executeCommand("set", envFile, "--app-key", "base64:MY_EXPLICIT_KEY=")
	require.NoError(t, err, "out=%s", out3)

	outAppKey3, err := executeCommand("get", envFile, "--app-key")
	require.NoError(t, err, "out=%s", outAppKey3)
	assert.Equal(t, "base64:MY_EXPLICIT_KEY=\n", outAppKey3)
}

func TestCLIEnableAndDisable(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// 1. Enable on non-existent file (should create file and set value to "true")
	out, err := executeCommand("enable", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Updated")
	assert.Contains(t, out, "APP_DEBUG")
	assert.Contains(t, out, `"true"`)

	out, err = executeCommand("get", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "true\n", out)

	// 2. Disable on existing "true" value (should change to "false" without confirmation)
	out, err = executeCommand("disable", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Updated")
	assert.Contains(t, out, "APP_DEBUG")
	assert.Contains(t, out, `"false"`)

	out, err = executeCommand("get", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "false\n", out)

	// 3. Enable on existing "false" value (should change to "true" without confirmation)
	out, err = executeCommand("enable", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_DEBUG")
	assert.Contains(t, out, `"true"`)

	// 4. Test with empty and null values (should not prompt)
	err = os.WriteFile(envFile, []byte("EMPTY_VAR=\nNULL_VAR=null\nUPPER_NULL=NULL\n"), 0644)
	require.NoError(t, err)

	out, err = executeCommand("enable", envFile, "--empty-var", "--null-var")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "EMPTY_VAR")
	assert.Contains(t, out, "NULL_VAR")
	assert.Contains(t, out, `"true"`)

	out, err = executeCommand("disable", envFile, "--upper-null")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "UPPER_NULL")
	assert.Contains(t, out, `"false"`)
}

func TestCLIEnableDisableConfirmation(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// Set initial non-boolean value
	err := os.WriteFile(envFile, []byte("APP_ENV=production\nDB_PORT=3306\n"), 0644)
	require.NoError(t, err)

	// Decline confirmation ("n")
	out, err := executeCommandWithInput("n\n", "enable", envFile, "--app-env")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "The environment variable 'APP_ENV' currently has value 'production'. Are you sure you want to change it to 'true'? (y/N):")
	assert.Contains(t, out, "No variables were updated.")

	// Verify APP_ENV is still production
	out, err = executeCommand("get", envFile, "--app-env")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "production\n", out)

	// Accept confirmation ("y")
	out, err = executeCommandWithInput("y\n", "enable", envFile, "--app-env")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "The environment variable 'APP_ENV' currently has value 'production'. Are you sure you want to change it to 'true'? (y/N):")
	assert.Contains(t, out, "APP_ENV")
	assert.Contains(t, out, `"true"`)

	out, err = executeCommand("get", envFile, "--app-env")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "true\n", out)

	// Test disable confirmation with "yes"
	out, err = executeCommandWithInput("yes\n", "disable", envFile, "--db-port")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "The environment variable 'DB_PORT' currently has value '3306'. Are you sure you want to change it to 'false'? (y/N):")
	assert.Contains(t, out, "DB_PORT")
	assert.Contains(t, out, `"false"`)

	out, err = executeCommand("get", envFile, "--db-port")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "false\n", out)
}

func TestCLIDotenvEnableDisable(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	out, err := executeCommand("dotenv", "enable", envFile, "--feature-flag")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "FEATURE_FLAG")
	assert.Contains(t, out, `"true"`)

	out, err = executeCommand("dotenv", "disable", envFile, "--feature-flag")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "FEATURE_FLAG")
	assert.Contains(t, out, `"false"`)
}

func TestFindExampleFile(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")
	exampleFile := filepath.Join(tempDir, ".env.example")
	envLocalFile := filepath.Join(tempDir, ".env.local")
	envLocalExampleFile := filepath.Join(tempDir, ".env.local.example")

	// No example files yet
	assert.Empty(t, FindExampleFile(envFile))
	assert.Empty(t, FindExampleFile(envLocalFile))

	// Create .env.example
	err := os.WriteFile(exampleFile, []byte("FOO=bar\n"), 0644)
	require.NoError(t, err)

	assert.Equal(t, exampleFile, FindExampleFile(envFile))
	// For .env.local, should fallback to .env.example if .env.local.example doesn't exist
	assert.Equal(t, exampleFile, FindExampleFile(envLocalFile))

	// Create .env.local.example
	err = os.WriteFile(envLocalExampleFile, []byte("LOCAL=true\n"), 0644)
	require.NoError(t, err)

	// Now .env.local should find .env.local.example
	assert.Equal(t, envLocalExampleFile, FindExampleFile(envLocalFile))
}

func TestCLIUseExampleFileOnSet(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")
	exampleFile := filepath.Join(tempDir, ".env.example")

	exampleContent := `# App Configuration
APP_NAME="Template App"
APP_ENV=local
DB_HOST=127.0.0.1
`
	err := os.WriteFile(exampleFile, []byte(exampleContent), 0644)
	require.NoError(t, err)

	// 1. Decline using example file ("n")
	out, err := executeCommandWithInput("n\n", "set", envFile, "--app-name", "CustomApp")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("The file '%s' does not exist. Would you like to use '%s'? (y/N):", envFile, exampleFile))
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"CustomApp"`)

	// APP_NAME should exist, but DB_HOST should NOT exist because template was declined
	outName, err := executeCommand("get", envFile, "--app-name")
	require.NoError(t, err, "out=%s", outName)
	assert.Equal(t, "CustomApp\n", outName)

	_, err = executeCommand("get", envFile, "--db-host")
	require.Error(t, err)

	// Clean up .env for next test
	_ = os.Remove(envFile)

	// 2. Accept using example file ("y")
	out, err = executeCommandWithInput("y\n", "set", envFile, "--app-name", "CustomApp")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("The file '%s' does not exist. Would you like to use '%s'? (y/N):", envFile, exampleFile))
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"CustomApp"`)

	// Verify both APP_NAME (overridden) and DB_HOST & APP_ENV (from template) exist
	outName, err = executeCommand("get", envFile, "--app-name")
	require.NoError(t, err, "out=%s", outName)
	assert.Equal(t, "CustomApp\n", outName)

	outEnv, err := executeCommand("get", envFile, "--app-env")
	require.NoError(t, err, "out=%s", outEnv)
	assert.Equal(t, "local\n", outEnv)

	outHost, err := executeCommand("get", envFile, "--db-host")
	require.NoError(t, err, "out=%s", outHost)
	assert.Equal(t, "127.0.0.1\n", outHost)

	// Verify comments from .env.example were preserved
	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# App Configuration")
}

func TestCLIUseExampleFileOnEnableDisable(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")
	exampleFile := filepath.Join(tempDir, ".env.example")

	exampleContent := `APP_DEBUG=false
APP_ENV=production
`
	err := os.WriteFile(exampleFile, []byte(exampleContent), 0644)
	require.NoError(t, err)

	// Accept using example file ("yes") when enabling APP_DEBUG
	out, err := executeCommandWithInput("yes\n", "enable", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("The file '%s' does not exist. Would you like to use '%s'? (y/N):", envFile, exampleFile))
	assert.Contains(t, out, "APP_DEBUG")
	assert.Contains(t, out, `"true"`)

	outDebug, err := executeCommand("get", envFile, "--app-debug")
	require.NoError(t, err, "out=%s", outDebug)
	assert.Equal(t, "true\n", outDebug)

	outEnv, err := executeCommand("get", envFile, "--app-env")
	require.NoError(t, err, "out=%s", outEnv)
	assert.Equal(t, "production\n", outEnv)
}

func TestCLICommentsOnSet(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// 1. Set with comment="..." prefix before subcommand
	out, err := executeCommand(`comment="Production API Key"`, "set", envFile, "--app-key=@generate")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Updated")
	assert.Contains(t, out, "APP_KEY")

	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Production API Key\nAPP_KEY=")

	// 2. Set with -c flag
	out, err = executeCommand("set", envFile, "--app-name", "Neon Nights", "-c", "Main application name")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"Neon Nights"`)
	assert.Contains(t, out, "Main application name")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Main application name\nAPP_NAME=\"Neon Nights\"")

	// 2b. Set with -c=... flag and --flag=... syntax
	out, err = executeCommand("set", envFile, "-c=Equals comment", "--app-desc=Testing description")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_DESC")
	assert.Contains(t, out, `"Testing description"`)
	assert.Contains(t, out, "Equals comment")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Equals comment\nAPP_DESC=\"Testing description\"")

	// 2c. Set with -c before subcommand
	out, err = executeCommand("-c", "Prefix short comment", "set", envFile, "--app-author=Adam")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_AUTHOR")
	assert.Contains(t, out, `"Adam"`)
	assert.Contains(t, out, "Prefix short comment")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Prefix short comment\nAPP_AUTHOR=\"Adam\"")

	// 3. Set with --comment-text flag
	out, err = executeCommand("set", envFile, "--db-port", "3306", "--comment-text=Database port")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "DB_PORT")
	assert.Contains(t, out, `"3306"`)
	assert.Contains(t, out, "Database port")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Database port\nDB_PORT=\"3306\"")

	// 4. Set with --with-comment flag
	out, err = executeCommand("set", envFile, "--redis-host", "127.0.0.1", "--with-comment=Redis configuration")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "REDIS_HOST")
	assert.Contains(t, out, `"127.0.0.1"`)
	assert.Contains(t, out, "Redis configuration")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Redis configuration\nREDIS_HOST=\"127.0.0.1\"")

	// 5. Set with --comment flag alongside other variable
	out, err = executeCommand("set", envFile, "--statamic-key", "secret123", "--comment=Statamic license key")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "STATAMIC_KEY")
	assert.Contains(t, out, `"********"`)
	assert.Contains(t, out, "Statamic license key")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Statamic license key\nSTATAMIC_KEY=\"secret123\"")

	// 6. Set with positional comment="..." argument
	out, err = executeCommand("set", envFile, "comment=Positional comment", "--queue-driver", "redis")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "QUEUE_DRIVER")
	assert.Contains(t, out, `"redis"`)
	assert.Contains(t, out, "Positional comment")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Positional comment\nQUEUE_DRIVER=\"redis\"")

	// 7. Set with a flag that has no explicit value, only a comment (existing value preserved)
	out, err = executeCommand("set", envFile, "--app-name", "-c", "Adds a comment, but won't overwrite existing value")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"Neon Nights"`)
	assert.Contains(t, out, "Adds a comment, but won't overwrite existing value")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Adds a comment, but won't overwrite existing value\nAPP_NAME=\"Neon Nights\"")

	// 8. Regression: set without a comment flag should not include any comment markers in the output
	out, err = executeCommand("set", envFile, "--app-timezone", "UTC")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_TIMEZONE")
	assert.Contains(t, out, `"UTC"`)
}

func TestCLIEnvVariableNamedComment(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// 1. Setting an environment variable named COMMENT without dedicated comment flag
	out, err := executeCommand("set", envFile, "--comment", "This is an env variable value")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "COMMENT")
	assert.Contains(t, out, `"This is an env variable value"`)

	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Equal(t, "COMMENT=\"This is an env variable value\"\n", string(content))

	// Getting the COMMENT environment variable
	out, err = executeCommand("get", envFile, "--comment")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "This is an env variable value\n", out)

	// 2. Setting COMMENT variable with a comment using -c
	out, err = executeCommand("set", envFile, "--comment", "New comment value", "-c", "Comment for COMMENT variable")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "COMMENT")
	assert.Contains(t, out, `"New comment value"`)
	assert.Contains(t, out, "Comment for COMMENT variable")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Comment for COMMENT variable\nCOMMENT=\"New comment value\"")

	// Getting COMMENT variable with --with-comments
	out, err = executeCommand("get", envFile, "--comment", "--with-comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "COMMENT")
	assert.Contains(t, out, "New comment value")
	assert.Contains(t, out, "Comment for COMMENT variable")

	// Getting COMMENT variable with -c
	out, err = executeCommand("get", envFile, "--comment", "-c")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "COMMENT")
	assert.Contains(t, out, "New comment value")
	assert.Contains(t, out, "Comment for COMMENT variable")

	// 3. Setting COMMENT variable with comment="..." prefix
	out, err = executeCommand(`comment="Prefix comment"`, "set", envFile, "--comment", "Value 3")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "COMMENT")
	assert.Contains(t, out, `"Value 3"`)

	out, err = executeCommand("get", envFile, "--comment", "--comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "COMMENT")
	assert.Contains(t, out, "Value 3")
	assert.Contains(t, out, "Prefix comment")
}

func TestCLIGetWithComments(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	initialContent := `# Application Configuration

# Production API Key
APP_KEY="base64:generatedkey123"

# Database Port
DB_PORT="3306"

# Multi-line line 1
# Multi-line line 2
MULTI_VAR="multi_value"

NO_COMMENT_VAR="plain_value"
`
	err := os.WriteFile(envFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	// 1. Standard get (no comments)
	out, err := executeCommand("get", envFile, "--app-key")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "base64:generatedkey123\n", out)

	// 2. Get with --with-comments
	out, err = executeCommand("get", envFile, "--app-key", "--with-comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "APP_KEY")
	assert.Contains(t, out, "base64:generatedkey123")
	assert.Contains(t, out, "Production API Key")

	// 3. Get with -c
	out, err = executeCommand("get", envFile, "--app-key", "-c")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "APP_KEY")
	assert.Contains(t, out, "base64:generatedkey123")
	assert.Contains(t, out, "Production API Key")

	// 4. Get with --comments
	out, err = executeCommand("get", envFile, "--db-port", "--comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "DB_PORT")
	assert.Contains(t, out, "3306")
	assert.Contains(t, out, "Database Port")

	// 5. Get with --with-comment
	out, err = executeCommand("get", envFile, "--db-port", "--with-comment")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "DB_PORT")
	assert.Contains(t, out, "3306")
	assert.Contains(t, out, "Database Port")

	// 6. Get multi-line comments
	out, err = executeCommand("get", envFile, "--multi-var", "--with-comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "MULTI_VAR")
	assert.Contains(t, out, "multi_value")
	assert.Contains(t, out, "Multi-line line 1")

	// 7. Get variable without comments using --with-comments flag
	out, err = executeCommand("get", envFile, "--no-comment-var", "--with-comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "NO_COMMENT_VAR")
	assert.Contains(t, out, "plain_value")
}

func TestCLIDotenvComments(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// Set with DDEV syntax and comment
	out, err := executeCommand("dotenv", "set", envFile, "--app-name", "DotEnvApp", "-c", "DDEV App Name")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"DotEnvApp"`)
	assert.Contains(t, out, "DDEV App Name")

	// Get with DDEV syntax and comments flag
	out, err = executeCommand("dotenv", "get", envFile, "--app-name", "--with-comments")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, "DotEnvApp")
	assert.Contains(t, out, "DDEV App Name")
}

func TestCLIYesFlag(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")
	exampleFile := filepath.Join(tempDir, ".env.example")

	exampleContent := `# App Config
APP_NAME="Template App"
APP_ENV=local
DB_HOST=127.0.0.1
`
	err := os.WriteFile(exampleFile, []byte(exampleContent), 0644)
	require.NoError(t, err)

	// 1. Set with -y flag when .env does not exist -> uses .env.example without prompting
	out, err := executeCommand("set", envFile, "--app-name", "New App", "-y")
	require.NoError(t, err, "out=%s", out)
	assert.NotContains(t, out, "Would you like to use")
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"New App"`)

	// Verify template values DB_HOST and APP_ENV were initialized
	outHost, err := executeCommand("get", envFile, "--db-host")
	require.NoError(t, err, "out=%s", outHost)
	assert.Equal(t, "127.0.0.1\n", outHost)

	// Clean up .env
	_ = os.Remove(envFile)

	// 2. Set with --yes flag
	out, err = executeCommand("set", envFile, "--app-name", "New App 2", "--yes")
	require.NoError(t, err, "out=%s", out)
	assert.NotContains(t, out, "Would you like to use")
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"New App 2"`)

	// Clean up .env
	_ = os.Remove(envFile)

	// 3. Dotenv set with -y flag
	out, err = executeCommand("dotenv", "set", envFile, "--app-name", "DDEV App", "-y")
	require.NoError(t, err, "out=%s", out)
	assert.NotContains(t, out, "Would you like to use")
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"DDEV App"`)

	// 4. Overwrite non-standard value in enable with -y (skips confirmation prompt)
	err = os.WriteFile(envFile, []byte("APP_DEBUG=custom_debug_val\n"), 0644)
	require.NoError(t, err)

	out, err = executeCommand("enable", envFile, "--app-debug", "-y")
	require.NoError(t, err, "out=%s", out)
	assert.NotContains(t, out, "Are you sure you want to change it")
	assert.Contains(t, out, "APP_DEBUG")
	assert.Contains(t, out, `"true"`)

	// 5. Overwrite non-standard value in disable with -y (skips confirmation prompt)
	err = os.WriteFile(envFile, []byte("APP_DEBUG=custom_debug_val\n"), 0644)
	require.NoError(t, err)

	out, err = executeCommand("disable", envFile, "--app-debug", "-y")
	require.NoError(t, err, "out=%s", out)
	assert.NotContains(t, out, "Are you sure you want to change it")
	assert.Contains(t, out, "APP_DEBUG")
	assert.Contains(t, out, `"false"`)
}

func TestCLISetCommentWithoutSettingVariable(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	initialContent := `# Application Config
APP_NAME="Original Name"
APP_ENV=local
DB_HOST=127.0.0.1
`
	err := os.WriteFile(envFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	// 1. Direct invocation on root command to set comment without setting variable value: ./gotenv .env --app-env -c "App env"
	out, err := executeCommand(envFile, "--app-env", "-c", "App env")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_ENV")
	assert.Contains(t, out, `"local"`)

	// Verify file content: APP_ENV value remained "local" and comment was added
	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# App env\nAPP_ENV=\"local\"")
	assert.Contains(t, string(content), "APP_NAME=\"Original Name\"")

	// 2. Set subcommand to update comment without setting variable: ./gotenv set .env --app-name -c "Application Title"
	out, err = executeCommand("set", envFile, "--app-name", "-c", "Application Title")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"Original Name"`)

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Application Title\nAPP_NAME=\"Original Name\"")

	// 3. Prefix comment syntax with direct invocation: ./gotenv comment="Database Host" .env --db-host
	out, err = executeCommand(`comment="Database Host"`, envFile, "--db-host")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "DB_HOST")
	assert.Contains(t, out, `"127.0.0.1"`)

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Database Host\nDB_HOST=\"127.0.0.1\"")

	// 4. Adding comment for a non-existing variable without value creates empty variable with comment
	out, err = executeCommand(envFile, "--new-feature", "-c", "New feature toggle")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "NEW_FEATURE")
	assert.Contains(t, out, `""`)

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# New feature toggle\nNEW_FEATURE=\"\"")

	// 5. Direct root invocation setting variable value: ./gotenv .env --app-name="New Direct Name"
	out, err = executeCommand(envFile, "--app-name", "New Direct Name")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"New Direct Name"`)

	// 6. Dotenv direct invocation: ./gotenv dotenv .env --app-env -c "Updated DDEV App Env"
	out, err = executeCommand("dotenv", envFile, "--app-env", "-c", "Updated DDEV App Env")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "APP_ENV")
	assert.Contains(t, out, `"local"`)

	// 7. Retrieve with comments via get: ./gotenv get .env --app-env -c
	out, err = executeCommand("get", envFile, "--app-env", "-c")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "KEY")
	assert.Contains(t, out, "APP_ENV")
	assert.Contains(t, out, "local")
	assert.Contains(t, out, "Updated DDEV App Env")
}

func TestCLISecretObfuscation(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	// 1. Setting variables where one is non-secret and others are secrets
	out, err := executeCommand("set", envFile,
		"--app-name", "My Secure App",
		"--app-key", "base64:uniqueproductionappkey12345678",
		"--db-password", "super_secret_db_pass",
		"--statamic-license-key", "statamic-secret-token",
	)
	require.NoError(t, err, "out=%s", out)

	// Console output should obfuscate secrets with "********"
	assert.Contains(t, out, "APP_NAME")
	assert.Contains(t, out, `"My Secure App"`)
	assert.Contains(t, out, "APP_KEY")
	assert.Contains(t, out, `"********"`)
	assert.Contains(t, out, "DB_PASSWORD")
	assert.Contains(t, out, `"********"`)
	assert.Contains(t, out, "STATAMIC_LICENSE_KEY")
	assert.Contains(t, out, `"********"`)
	assert.NotContains(t, out, "uniqueproductionappkey12345678")
	assert.NotContains(t, out, "super_secret_db_pass")
	assert.NotContains(t, out, "statamic-secret-token")

	// The actual .env file on disk MUST have the real unabbreviated values
	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	contentStr := string(content)
	assert.Contains(t, contentStr, `APP_NAME="My Secure App"`)
	assert.Contains(t, contentStr, `APP_KEY="base64:uniqueproductionappkey12345678"`)
	assert.Contains(t, contentStr, `DB_PASSWORD="super_secret_db_pass"`)
	assert.Contains(t, contentStr, `STATAMIC_LICENSE_KEY="statamic-secret-token"`)

	// 2. Getting values via get command should return the unmasked secret value for scripting
	outKey, err := executeCommand("get", envFile, "--app-key")
	require.NoError(t, err, "out=%s", outKey)
	assert.Equal(t, "base64:uniqueproductionappkey12345678\n", outKey)

	outPass, err := executeCommand("get", envFile, "--db-password")
	require.NoError(t, err, "out=%s", outPass)
	assert.Equal(t, "super_secret_db_pass\n", outPass)

	// 3. Setting with generated keys (@generate) also obfuscates output
	outGen, err := executeCommand("set", envFile, "--secret-key", "@generate")
	require.NoError(t, err, "out=%s", outGen)
	assert.Contains(t, outGen, "SECRET_KEY")
	assert.Contains(t, outGen, `"********"`)

	outSecret, err := executeCommand("get", envFile, "--secret-key")
	require.NoError(t, err, "out=%s", outSecret)
	assert.NotEqual(t, "********\n", outSecret)
	assert.Len(t, strings.TrimSpace(outSecret), 32)
}

func TestCLIUnsetVsRm(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	initialContent := `# License key for Statamic
STATAMIC_LICENSE_KEY=secret123
APP_NAME=TestApp
`
	err := os.WriteFile(envFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	// unset should blank the value but keep the variable (and its comment) in place
	out, err := executeCommand("unset", envFile, "--statamic-license-key")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Unset 1 variable(s)")

	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# License key for Statamic")
	assert.Contains(t, string(content), "STATAMIC_LICENSE_KEY=")
	assert.NotContains(t, string(content), "secret123")

	out, err = executeCommand("get", envFile, "--statamic-license-key")
	require.NoError(t, err, "out=%s", out)
	assert.Equal(t, "\n", out)

	// rm/delete should remove the variable entirely
	out, err = executeCommand("delete", envFile, "--app-name")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Removed 1 variable(s)")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "APP_NAME")

	_, err = executeCommand("get", envFile, "--app-name")
	require.Error(t, err)
}

func TestCLIRmKeepsComments(t *testing.T) {
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	initialContent := `# License key for Statamic
STATAMIC_LICENSE_KEY=secret123
APP_NAME=TestApp
`
	err := os.WriteFile(envFile, []byte(initialContent), 0644)
	require.NoError(t, err)

	// rm should remove the variable line, but always leave any preceding comment untouched
	out, err := executeCommand("rm", envFile, "--statamic-license-key")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Removed 1 variable(s)")

	content, err := os.ReadFile(envFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# License key for Statamic")
	assert.NotContains(t, string(content), "STATAMIC_LICENSE_KEY")

	out, err = executeCommand("delete", envFile, "--app-name")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Removed 1 variable(s)")

	content, err = os.ReadFile(envFile)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "APP_NAME")
}

func TestCLISetWithAndFlag(t *testing.T) {
	tempDir := t.TempDir()
	cmsDir := filepath.Join(tempDir, "cms")
	webDir := filepath.Join(tempDir, "web")
	require.NoError(t, os.MkdirAll(cmsDir, 0755))
	require.NoError(t, os.MkdirAll(webDir, 0755))

	cmsEnv := filepath.Join(cmsDir, ".env")
	webEnv := filepath.Join(webDir, ".env.local")

	// Set value with @generate on two files simultaneously
	out, err := executeCommand("set", cmsEnv, fmt.Sprintf("--and=%s", webEnv), "--app-secret=@generate", "--shared-title=Monorepo App")
	require.NoError(t, err, "out=%s", out)

	assert.Contains(t, out, fmt.Sprintf("Updated %s", cmsEnv))
	assert.Contains(t, out, fmt.Sprintf("Updated %s", webEnv))
	assert.Contains(t, out, "APP_SECRET")
	assert.Contains(t, out, "SHARED_TITLE")

	// Both files must have the EXACT same generated APP_SECRET value
	outCmsSecret, err := executeCommand("get", cmsEnv, "--app-secret")
	require.NoError(t, err, "out=%s", outCmsSecret)

	outWebSecret, err := executeCommand("get", webEnv, "--app-secret")
	require.NoError(t, err, "out=%s", outWebSecret)

	assert.NotEmpty(t, strings.TrimSpace(outCmsSecret))
	assert.Equal(t, outCmsSecret, outWebSecret)

	// Both files must have SHARED_TITLE
	outCmsTitle, err := executeCommand("get", cmsEnv, "--shared-title")
	require.NoError(t, err, "out=%s", outCmsTitle)
	assert.Equal(t, "Monorepo App\n", outCmsTitle)

	outWebTitle, err := executeCommand("get", webEnv, "--shared-title")
	require.NoError(t, err, "out=%s", outWebTitle)
	assert.Equal(t, "Monorepo App\n", outWebTitle)
}

func TestCLIPushCommand(t *testing.T) {
	tempDir := t.TempDir()
	sourceEnv := filepath.Join(tempDir, "web", ".env.local")
	targetEnv := filepath.Join(tempDir, "cms", ".env")
	extraTargetEnv := filepath.Join(tempDir, "api", ".env")

	require.NoError(t, os.MkdirAll(filepath.Dir(sourceEnv), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(targetEnv), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(extraTargetEnv), 0755))

	sourceContent := `# Application Secret Key
APP_SECRET=source_super_secret_value_123
# Database Connection Info
DB_HOST=127.0.0.1
DB_PORT=5432
API_KEY=web_api_key_xyz
`
	require.NoError(t, os.WriteFile(sourceEnv, []byte(sourceContent), 0644))

	// 1. Push with --key, --from, and --to flags: gotenv push --key=APP_SECRET --from=web/.env.local --to=cms/.env
	out, err := executeCommand("push", "--key=APP_SECRET", fmt.Sprintf("--from=%s", sourceEnv), fmt.Sprintf("--to=%s", targetEnv))
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("Updated %s", targetEnv))
	assert.Contains(t, out, "APP_SECRET")

	outSecret, err := executeCommand("get", targetEnv, "--app-secret")
	require.NoError(t, err, "out=%s", outSecret)
	assert.Equal(t, "source_super_secret_value_123\n", outSecret)

	// Comment from source file should be preserved
	content, err := os.ReadFile(targetEnv)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Application Secret Key\nAPP_SECRET=\"source_super_secret_value_123\"")

	// 2. Push positional thing, from, to: gotenv push DB_HOST DB_PORT --from=... --to=... with comment override
	out, err = executeCommand("push", "DB_HOST", "DB_PORT", fmt.Sprintf("--from=%s", sourceEnv), fmt.Sprintf("--to=%s", targetEnv), "-c", "Synced DB config")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "DB_HOST")
	assert.Contains(t, out, "DB_PORT")

	outHost, err := executeCommand("get", targetEnv, "--db-host")
	require.NoError(t, err, "out=%s", outHost)
	assert.Equal(t, "127.0.0.1\n", outHost)

	content, err = os.ReadFile(targetEnv)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Synced DB config\nDB_HOST=\"127.0.0.1\"")

	// 3. Push thing, from, to, and: gotenv push API_KEY --from=... --to=... --and=...
	out, err = executeCommand("push", "API_KEY", fmt.Sprintf("--from=%s", sourceEnv), fmt.Sprintf("--to=%s", targetEnv), fmt.Sprintf("--and=%s", extraTargetEnv))
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("Updated %s", targetEnv))
	assert.Contains(t, out, fmt.Sprintf("Updated %s", extraTargetEnv))

	outApiTarget, err := executeCommand("get", targetEnv, "--api-key")
	require.NoError(t, err, "out=%s", outApiTarget)
	assert.Equal(t, "web_api_key_xyz\n", outApiTarget)

	outApiExtra, err := executeCommand("get", extraTargetEnv, "--api-key")
	require.NoError(t, err, "out=%s", outApiExtra)
	assert.Equal(t, "web_api_key_xyz\n", outApiExtra)

	// 4. Push --all flag
	out, err = executeCommand("push", fmt.Sprintf("--from=%s", sourceEnv), fmt.Sprintf("--to=%s", extraTargetEnv), "--all")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("Updated %s", extraTargetEnv))

	// 5. Backwards compatibility: sync alias with dynamic flag
	out, err = executeCommand("sync", targetEnv, fmt.Sprintf("--with=%s", sourceEnv), "--app-secret")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, fmt.Sprintf("Updated %s", targetEnv))

	// 6. Error when variable does not exist in source file
	_, err = executeCommand("push", "--key=NON_EXISTENT_KEY", fmt.Sprintf("--from=%s", sourceEnv), fmt.Sprintf("--to=%s", targetEnv))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NON_EXISTENT_KEY")

	// 7. Error when --from is missing
	_, err = executeCommand("push", "--key=APP_SECRET", fmt.Sprintf("--to=%s", targetEnv))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required source file flag --from")

	// 8. Error when source and target are the same file
	_, err = executeCommand("push", "--key=APP_SECRET", fmt.Sprintf("--from=%s", sourceEnv), fmt.Sprintf("--to=%s", sourceEnv))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "source file and target file cannot be the same")

	// 9. Relative monorepo paths like ../web/.env.local
	cmsDir := filepath.Join(tempDir, "cms_app")
	webDir := filepath.Join(tempDir, "web_app")
	require.NoError(t, os.MkdirAll(cmsDir, 0755))
	require.NoError(t, os.MkdirAll(webDir, 0755))

	cmsEnv := filepath.Join(cmsDir, ".env")
	webLocalEnv := filepath.Join(webDir, ".env.local")
	require.NoError(t, os.WriteFile(cmsEnv, []byte("REFRESH_TOKEN=cms_refresh_token_999\nAUTH_SECRET=auth_sec_abc\n"), 0644))

	origWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origWd)

	require.NoError(t, os.Chdir(cmsDir))

	// Test pushing from .env to ../web_app/.env.local
	out, err = executeCommand("push", "--key=REFRESH_TOKEN", "--from=.env", "--to=../web_app/.env.local")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "REFRESH_TOKEN")

	content, err = os.ReadFile(webLocalEnv)
	require.NoError(t, err)
	assert.Contains(t, string(content), `REFRESH_TOKEN="cms_refresh_token_999"`)

	// Test pushing with positional target and dynamic flag: gotenv push ../web_app/.env.local --from=.env --auth-secret
	out, err = executeCommand("push", "../web_app/.env.local", "--from=.env", "--auth-secret")
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "AUTH_SECRET")

	content, err = os.ReadFile(webLocalEnv)
	require.NoError(t, err)
	assert.Contains(t, string(content), `AUTH_SECRET="auth_sec_abc"`)

	// Restore working directory before test completion
	_ = os.Chdir(origWd)
}

func TestCLISkillCommand(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Generate SKILL.md with -out flag to directory
	skillDir := filepath.Join(tempDir, "agent-skills")
	out, err := executeCommand("skill", "-out", skillDir)
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Successfully generated")

	skillFile := filepath.Join(skillDir, "SKILL.md")
	require.FileExists(t, skillFile)

	content, err := os.ReadFile(skillFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: gotenv")
	assert.Contains(t, string(content), "description:")
	assert.Contains(t, string(content), "# gotenv Skill")
	assert.Contains(t, string(content), "gotenv [command] [file] [flags]")

	// 2. Generate with --out flag
	skillDir2 := filepath.Join(tempDir, "another-dir")
	out, err = executeCommand("skill", "--out", skillDir2)
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Successfully generated")
	require.FileExists(t, filepath.Join(skillDir2, "SKILL.md"))

	// 3. Generate with -o shorthand flag
	skillDir3 := filepath.Join(tempDir, "short-dir")
	out, err = executeCommand("skill", "-o", skillDir3)
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Successfully generated")
	require.FileExists(t, filepath.Join(skillDir3, "SKILL.md"))

	// 4. Generate with explicit .md file destination
	customMd := filepath.Join(tempDir, "docs", "CUSTOM_SKILL.md")
	out, err = executeCommand("skill", "-out", customMd)
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Successfully generated")
	require.FileExists(t, customMd)

	// 5. Generate with positional argument
	posDir := filepath.Join(tempDir, "pos-dir")
	out, err = executeCommand("skill", posDir)
	require.NoError(t, err, "out=%s", out)
	assert.Contains(t, out, "Successfully generated")
	require.FileExists(t, filepath.Join(posDir, "SKILL.md"))
}
