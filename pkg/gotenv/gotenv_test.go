package gotenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAndWriteEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	initialContent := `# Project Configuration
APP_NAME=OldApp
# APP_PORT=3000
DATABASE_URL="postgres://localhost:5432/db"
`
	err := os.WriteFile(envPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	envMap, envText, err := ReadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "OldApp", envMap["APP_NAME"])
	assert.Equal(t, "postgres://localhost:5432/db", envMap["DATABASE_URL"])
	assert.Contains(t, envText, "APP_NAME=OldApp")

	// Update existing and add new
	updates := map[string]string{
		"APP_NAME":   "NewApp",
		"APP_PORT":   "8080",
		"SECRET_KEY": `pass"word$with$special`,
		"MULTI_LINE": `line1\nline2`,
	}

	err = WriteEnvFile(envPath, updates, envText)
	require.NoError(t, err)

	newMap, newText, err := ReadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "NewApp", newMap["APP_NAME"])
	assert.Equal(t, "8080", newMap["APP_PORT"])
	assert.Equal(t, `pass"word$with$special`, newMap["SECRET_KEY"])
	assert.Equal(t, "line1\nline2", newMap["MULTI_LINE"])
	assert.Equal(t, `line1\nline2`, FormatEnvValue(newMap["MULTI_LINE"]))

	// Check comments and formatting preserved
	assert.Contains(t, newText, "# Project Configuration")
	assert.Contains(t, newText, `APP_NAME="NewApp"`)
	assert.Contains(t, newText, `APP_PORT="8080"`)
	assert.Contains(t, newText, `SECRET_KEY="pass\"word\$with\$special"`)
}

func TestRemoveEnvFileKeys(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	initialContent := `KEY1="val1"
KEY2="val2"
KEY3="val3"
`
	err := os.WriteFile(envPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	_, envText, err := ReadEnvFile(envPath)
	require.NoError(t, err)

	removed, err := RemoveEnvFileKeys(envPath, []string{"KEY2"}, envText)
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	newMap, newText, err := ReadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "val1", newMap["KEY1"])
	assert.Equal(t, "val3", newMap["KEY3"])
	assert.NotContains(t, newText, "KEY2")
}

func TestWriteToNewFile(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, "nested", "dir", ".env.local")

	updates := map[string]string{
		"ENV": "production",
	}

	err := WriteEnvFile(envPath, updates, "")
	require.NoError(t, err)

	envMap, envText, err := ReadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "production", envMap["ENV"])
	assert.Equal(t, "ENV=\"production\"\n", envText)
}

func TestComments(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")

	initialContent := `# Header Comment

# Old App Comment
APP_NAME=OldApp
# APP_PORT=3000

# Secret Key
APP_KEY="xyz"
`
	err := os.WriteFile(envPath, []byte(initialContent), 0644)
	require.NoError(t, err)

	_, envText, err := ReadEnvFile(envPath)
	require.NoError(t, err)

	// Test GetEnvVarComment
	assert.Equal(t, "# Old App Comment", GetEnvVarComment(envText, "APP_NAME"))
	assert.Equal(t, "# Secret Key", GetEnvVarComment(envText, "APP_KEY"))
	assert.Equal(t, "", GetEnvVarComment(envText, "NON_EXISTENT"))

	// Test FormatComment
	assert.Equal(t, "# Production Key\n", FormatComment("Production Key"))
	assert.Equal(t, "# Production Key\n", FormatComment("# Production Key"))
	assert.Equal(t, "# Line 1\n# Line 2\n", FormatComment("Line 1\nLine 2"))
	assert.Equal(t, "# Line 1\n# Line 2\n", FormatComment("# Line 1\n# Line 2"))
	assert.Equal(t, "", FormatComment(""))

	// Test WriteEnvFileWithComments - update existing key with new comment
	updates := map[string]string{
		"APP_NAME": "NewApp",
		"NEW_VAR":  "new_val",
	}
	comments := map[string]string{
		"APP_NAME": "Updated App Name",
		"NEW_VAR":  "Brand New Variable",
	}

	err = WriteEnvFileWithComments(envPath, updates, comments, envText)
	require.NoError(t, err)

	newMap, newText, err := ReadEnvFile(envPath)
	require.NoError(t, err)
	assert.Equal(t, "NewApp", newMap["APP_NAME"])
	assert.Equal(t, "new_val", newMap["NEW_VAR"])
	assert.Equal(t, "# Updated App Name", GetEnvVarComment(newText, "APP_NAME"))
	assert.Equal(t, "# Brand New Variable", GetEnvVarComment(newText, "NEW_VAR"))
	assert.Equal(t, "# Secret Key", GetEnvVarComment(newText, "APP_KEY"))
	assert.Contains(t, newText, "# Header Comment")
}

func TestIsSecretAndObfuscate(t *testing.T) {
	// Secret key names
	secretKeys := []string{
		"APP_KEY",
		"API_KEY",
		"APIKEY",
		"SECRET_KEY",
		"MY_SECRET",
		"CLIENT_SECRET",
		"DB_PASSWORD",
		"MYSQL_ROOT_PASSWORD",
		"REDIS_PASS",
		"AUTH_TOKEN",
		"ACCESS_TOKEN",
		"GITHUB_TOKEN",
		"AWS_CREDENTIALS",
		"PRIVATE_KEY",
		"SSL_CERT",
		"TLS_CERTIFICATE",
		"HMAC_SIGNATURE",
		"PASSWORD_SALT",
		"USER_PIN",
		"SENTRY_DSN",
		"STRIPE_WEBHOOK",
		"HTTP_AUTH",
		"BASIC_AUTH",
		"STATAMIC_LICENSE_KEY",
	}

	for _, key := range secretKeys {
		assert.True(t, IsSecret(key, "some-secret-value"), "Expected key %s to be secret", key)
	}

	// Non-secret key names
	nonSecretKeys := []string{
		"APP_NAME",
		"APP_URL",
		"APP_ENV",
		"APP_DEBUG",
		"APP_AUTHOR",
		"DB_PORT",
		"DB_HOST",
		"DB_DATABASE",
		"DB_USERNAME",
		"REDIS_HOST",
		"REDIS_PORT",
		"QUEUE_DRIVER",
		"FEATURE_FLAG",
		"COMMENT",
	}

	for _, key := range nonSecretKeys {
		assert.False(t, IsSecret(key, "standard_value"), "Expected key %s to not be secret", key)
	}

	// Secret detection by value even with non-descript key
	assert.True(t, IsSecret("CUSTOM_VAR", "base64:32byteencodedcharacters1234567890"))
	assert.True(t, IsSecret("CUSTOM_VAR", "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASC..."))
	assert.True(t, IsSecret("CUSTOM_VAR", "ghp_1234567890abcdefghijklmnopqrstuvwxyz"))
	assert.True(t, IsSecret("CUSTOM_VAR", "sk_live_1234567890abcdef"))

	// Empty and boolean values should not trigger secret masking
	assert.False(t, IsSecret("DB_PASSWORD", ""))
	assert.False(t, IsSecret("APP_KEY", "true"))
	assert.False(t, IsSecret("SECRET_KEY", "false"))
	assert.False(t, IsSecret("SECRET_KEY", "null"))

	// ObfuscateValue
	assert.Equal(t, "********", ObfuscateValue("my_super_secret"))
}
