package gotenv

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

var (
	varLineRegex      = regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=`)
	disabledLineRegex = regexp.MustCompile(`^#+\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=`)
)

func parseEnvLine(line string) (key string, isVar bool, isDisabled bool) {
	trimmed := strings.TrimSpace(line)
	if m := varLineRegex.FindStringSubmatch(trimmed); len(m) > 1 {
		return m[1], true, false
	}
	if m := disabledLineRegex.FindStringSubmatch(trimmed); len(m) > 1 {
		return m[1], true, true
	}
	return "", false, false
}

// ReadEnvFile reads a .env file and returns a map of key-value pairs,
// the unaltered plain text of the file, and any error encountered.
func ReadEnvFile(filePath string) (map[string]string, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}
	envText := string(data)

	envMap, err := godotenv.Unmarshal(envText)
	if err != nil {
		// Fallback to godotenv.Read in case of parsing nuances
		envMap, err = godotenv.Read(filePath)
		if err != nil {
			return nil, envText, err
		}
	}

	return envMap, envText, nil
}

// WriteEnvFileWithComments writes updated key-value pairs and optional comments into the specified .env file,
// preserving existing formatting and comments in envText where possible.
func WriteEnvFileWithComments(filePath string, envMap map[string]string, commentsMap map[string]string, envText string) error {
	var resultLines []string
	var pendingComments []string
	updatedKeys := make(map[string]bool)

	flushComments := func() {
		for _, c := range pendingComments {
			resultLines = append(resultLines, c)
		}
		pendingComments = nil
	}

	hasContent := strings.TrimSpace(envText) != ""

	if hasContent {
		lines := strings.Split(strings.ReplaceAll(envText, "\r\n", "\n"), "\n")
		// If last line is empty due to trailing newline, remove it for processing
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				flushComments()
				resultLines = append(resultLines, "")
				continue
			}

			key, isVar, isDisabled := parseEnvLine(line)
			if isVar {
				if val, ok := envMap[key]; ok {
					updatedKeys[key] = true
					comment := ""
					if commentsMap != nil {
						comment = commentsMap[key]
					}
					if strings.TrimSpace(comment) != "" {
						pendingComments = nil
						formattedComment := strings.TrimSuffix(FormatComment(comment), "\n")
						for _, fcl := range strings.Split(formattedComment, "\n") {
							resultLines = append(resultLines, fcl)
						}
					} else {
						flushComments()
					}
					resultLines = append(resultLines, fmt.Sprintf("%s=%s", key, EscapeEnvFileValue(val)))
					continue
				} else if isDisabled {
					pendingComments = append(pendingComments, line)
					continue
				} else {
					flushComments()
					resultLines = append(resultLines, line)
					continue
				}
			}

			if strings.HasPrefix(trimmed, "#") {
				pendingComments = append(pendingComments, line)
			} else {
				flushComments()
				resultLines = append(resultLines, line)
			}
		}

		flushComments()
	}

	// Append any new keys that were not present in envText
	var newKeys []string
	for k := range envMap {
		if !updatedKeys[k] {
			newKeys = append(newKeys, k)
		}
	}
	sort.Strings(newKeys)

	for _, k := range newKeys {
		comment := ""
		if commentsMap != nil {
			comment = commentsMap[k]
		}
		if strings.TrimSpace(comment) != "" {
			formattedComment := strings.TrimSuffix(FormatComment(comment), "\n")
			for _, fcl := range strings.Split(formattedComment, "\n") {
				resultLines = append(resultLines, fcl)
			}
		}
		resultLines = append(resultLines, fmt.Sprintf("%s=%s", k, EscapeEnvFileValue(envMap[k])))
	}

	output := strings.Join(resultLines, "\n")
	if output != "" {
		output += "\n"
	}

	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, []byte(output), 0644)
}

// WriteEnvFile writes updated key-value pairs into the specified .env file,
// preserving existing formatting and comments in envText where possible.
func WriteEnvFile(filePath string, envMap map[string]string, envText string) error {
	return WriteEnvFileWithComments(filePath, envMap, nil, envText)
}

// FormatComment formats a comment string into # comment lines ending with a newline.
func FormatComment(comment string) string {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return ""
	}
	lines := strings.Split(comment, "\n")
	var formatted []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			formatted = append(formatted, "#")
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			formatted = append(formatted, trimmed)
		} else {
			formatted = append(formatted, "# "+trimmed)
		}
	}
	return strings.Join(formatted, "\n") + "\n"
}

// GetEnvVarComment returns the comment lines directly preceding the given key in envText.
func GetEnvVarComment(envText, targetKey string) string {
	lines := strings.Split(strings.ReplaceAll(envText, "\r\n", "\n"), "\n")
	var currentComments []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			currentComments = nil
			continue
		}

		key, isVar, isDisabled := parseEnvLine(line)
		if isVar && !isDisabled {
			if key == targetKey {
				return strings.Join(currentComments, "\n")
			}
			currentComments = nil
		} else if isVar && isDisabled {
			if key == targetKey {
				return strings.Join(currentComments, "\n")
			}
			currentComments = append(currentComments, trimmed)
		} else if strings.HasPrefix(trimmed, "#") {
			currentComments = append(currentComments, trimmed)
		} else {
			currentComments = nil
		}
	}

	return ""
}

// BlankEnvFileKeys clears the value of specified keys, keeping the variable (and its comments) in place.
func BlankEnvFileKeys(filePath string, keys []string, envText string) (int, error) {
	blankedCount := 0
	for _, k := range keys {
		exp := regexp.MustCompile(fmt.Sprintf(`(?m)^([ \t]*%s[ \t]*=)(.*)$`, regexp.QuoteMeta(k)))
		if exp.MatchString(envText) {
			envText = exp.ReplaceAllString(envText, "$1")
			blankedCount++
		}
	}

	if err := os.WriteFile(filePath, []byte(envText), 0644); err != nil {
		return blankedCount, err
	}

	return blankedCount, nil
}

// RemoveEnvFileKeys removes specified keys from the .env content and writes it to file.
// Preceding comment lines are left untouched.
func RemoveEnvFileKeys(filePath string, keys []string, envText string) (int, error) {
	removedCount := 0
	for _, k := range keys {
		exp := regexp.MustCompile(fmt.Sprintf(`(^|[\r\n]+)#*[ \t]*(%s)[ \t]*=(.*)`, regexp.QuoteMeta(k)))
		if exp.MatchString(envText) {
			// Remove the entire line
			envText = exp.ReplaceAllString(envText, "")
			removedCount++
		}
	}

	// Normalize excessive blank newlines resulting from removal
	envText = strings.TrimSpace(envText)
	if envText != "" {
		envText += "\n"
	}

	if err := os.WriteFile(filePath, []byte(envText), 0644); err != nil {
		return removedCount, err
	}

	return removedCount, nil
}

// EscapeEnvFileValue escapes special characters in value and wraps it in double quotes.
func EscapeEnvFileValue(value string) string {
	escaped := strings.NewReplacer(
		// Escape all dollar signs so they are not interpreted as shell variables
		`$`, `\$`,
		// Escape double quotes
		`"`, `\"`,
	).Replace(value)

	return `"` + escaped + `"`
}

// FormatEnvValue returns the unquoted and unescaped representation of the variable value for output.
func FormatEnvValue(val string) string {
	return strings.Trim(fmt.Sprintf("%#v", val), `"`)
}

// ObfuscatedSecretMask is the replacement string used to hide secrets in console output.
const ObfuscatedSecretMask = "********"

// IsSecret determines if a variable name or its value looks like it would contain a secret.
func IsSecret(key string, val ...string) bool {
	upperKey := strings.ToUpper(strings.TrimSpace(key))
	upperKey = strings.ReplaceAll(upperKey, "-", "_")

	// If value is provided and is a trivial non-secret (empty, boolean, or null), do not treat as secret
	if len(val) > 0 {
		v := strings.TrimSpace(val[0])
		if v == "" || strings.EqualFold(v, "true") || strings.EqualFold(v, "false") || strings.EqualFold(v, "null") {
			return false
		}

		// Check value-based secret patterns
		upperVal := strings.ToUpper(v)
		if strings.HasPrefix(v, "base64:") && len(v) >= 20 {
			return true
		}
		if strings.Contains(upperVal, "BEGIN PRIVATE KEY") ||
			strings.Contains(upperVal, "BEGIN RSA PRIVATE KEY") ||
			strings.Contains(upperVal, "BEGIN OPENSSH PRIVATE KEY") ||
			strings.Contains(upperVal, "BEGIN EC PRIVATE KEY") ||
			strings.Contains(upperVal, "BEGIN CERTIFICATE") ||
			strings.Contains(upperVal, "BEGIN PGP PRIVATE KEY") {
			return true
		}
		if strings.HasPrefix(v, "ghp_") || strings.HasPrefix(v, "gho_") ||
			strings.HasPrefix(v, "github_pat_") || strings.HasPrefix(v, "glpat-") ||
			strings.HasPrefix(v, "xoxb-") || strings.HasPrefix(v, "xoxp-") ||
			strings.HasPrefix(v, "sk_live_") || strings.HasPrefix(v, "rk_live_") {
			return true
		}
	}

	// Direct substring keywords
	secretSubstrings := []string{
		"SECRET",
		"PASSWORD",
		"PASSWD",
		"TOKEN",
		"KEY",
		"CREDENTIAL",
		"PRIVATE",
		"SIGNATURE",
		"SALT",
		"WEBHOOK",
		"AUTHENTICAT",
		"AUTHORIZ",
		"OAUTH",
	}

	for _, sub := range secretSubstrings {
		if strings.Contains(upperKey, sub) {
			return true
		}
	}

	// Tokenized word checks (separated by underscore)
	parts := strings.Split(upperKey, "_")
	for _, p := range parts {
		if p == "PASS" || p == "AUTH" || p == "CERT" || p == "CERTIFICATE" || p == "PIN" || p == "DSN" {
			return true
		}
	}

	return false
}

// ObfuscateValue returns an obfuscated placeholder string to protect secrets in console output.
func ObfuscateValue(val string) string {
	return ObfuscatedSecretMask
}
