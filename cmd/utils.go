package cmd

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagRegex validates long and short flag formats:
// e.g. -a, -a=val, --app-key, --app-key=val
var FlagRegex = regexp.MustCompile(`^-[a-z]$|^-[a-z]=.*|^--[a-z][a-z0-9-]*$|^--[a-z][a-z0-9-]*=.*`)

var yesFlag bool

func findKnownFlag(cmd *cobra.Command, flagPart string) *pflag.Flag {
	if strings.HasPrefix(flagPart, "--") {
		name := strings.TrimPrefix(flagPart, "--")
		return cmd.Flags().Lookup(name)
	} else if strings.HasPrefix(flagPart, "-") {
		sh := strings.TrimPrefix(flagPart, "-")
		return cmd.Flags().ShorthandLookup(sh)
	}
	return nil
}

// ExtractCommentFromArgs scans args for comment specifications like comment="..." or -c "..." (when before subcommand)
// and returns the cleaned args slice and the extracted comment string.
func ExtractCommentFromArgs(args []string) ([]string, string) {
	var cleanArgs []string
	var comment string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		lower := strings.ToLower(arg)
		if strings.HasPrefix(lower, "comment=") {
			val := arg[8:]
			comment = strings.Trim(val, `"'`)
		} else if strings.HasPrefix(lower, "comment:") {
			val := arg[8:]
			comment = strings.Trim(val, `"'`)
		} else if (lower == "-c" || strings.HasPrefix(lower, "--comment-text") || strings.HasPrefix(lower, "--with-comment")) && len(cleanArgs) == 0 {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				comment = strings.Trim(parts[1], `"'`)
			} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				comment = strings.Trim(args[i+1], `"'`)
				i++
			} else {
				cleanArgs = append(cleanArgs, arg)
			}
		} else if arg == "-out" {
			cleanArgs = append(cleanArgs, "--out")
		} else if strings.HasPrefix(arg, "-out=") {
			cleanArgs = append(cleanArgs, "--out="+strings.TrimPrefix(arg, "-out="))
		} else if arg == "-output" {
			cleanArgs = append(cleanArgs, "--output")
		} else if strings.HasPrefix(arg, "-output=") {
			cleanArgs = append(cleanArgs, "--output="+strings.TrimPrefix(arg, "-output="))
		} else {
			cleanArgs = append(cleanArgs, arg)
		}
	}

	return cleanArgs, comment
}

// GetUnknownFlagsDetailed extracts flags that are not registered on the command,
// and returns both a map of flag to value and a map indicating whether each flag had an explicit value specified.
func GetUnknownFlagsDetailed(cmd *cobra.Command, customArgs ...[]string) (map[string]string, map[string]bool, error) {
	unknownFlags := make(map[string]string)
	hasExplicitVal := make(map[string]bool)

	var args []string
	if len(customArgs) > 0 && customArgs[0] != nil {
		args = customArgs[0]
	} else if len(os.Args) > 1 {
		args = os.Args[1:]
	} else {
		return unknownFlags, hasExplicitVal, nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Separate flag and inline value if "=" is present (e.g. --flag=value or -f=value)
		flagPart := arg
		hasInlineVal := false
		inlineVal := ""
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			flagPart = parts[0]
			inlineVal = parts[1]
			hasInlineVal = true
		}

		// Check if it's a known flag registered on the command or inherited from parent
		knownFlag := findKnownFlag(cmd, flagPart)
		if knownFlag != nil {
			if !hasInlineVal {
				// If the flag takes a value and not inline, skip the next argument
				takesValue := knownFlag.Value.Type() != "bool" && knownFlag.NoOptDefVal == ""
				if takesValue && i+1 < len(args) {
					i++
				}
			}
			continue
		}

		// If not starting with "-", it's a positional argument or subcommand name (e.g. "set", ".env")
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		// Validate flag format using FlagRegex
		if !FlagRegex.MatchString(arg) {
			return unknownFlags, hasExplicitVal, fmt.Errorf("the flag must consist of lowercase letters, numbers, and hyphens, and it must start with a letter, but received %s", arg)
		}

		// Handle unknown flag with value
		if hasInlineVal {
			unknownFlags[flagPart] = inlineVal
			hasExplicitVal[flagPart] = true
		} else if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			unknownFlags[flagPart] = args[i+1]
			hasExplicitVal[flagPart] = true
			i++
		} else {
			unknownFlags[flagPart] = ""
			hasExplicitVal[flagPart] = false
		}
	}

	return unknownFlags, hasExplicitVal, nil
}

// GetUnknownFlags extracts flags that are not registered on the command.
// It works on a provided slice of args or falls back to os.Args[1:].
func GetUnknownFlags(cmd *cobra.Command, customArgs ...[]string) (map[string]string, error) {
	flags, _, err := GetUnknownFlagsDetailed(cmd, customArgs...)
	return flags, err
}

// FlagToEnvKey converts a flag like "--app-key" or "--db-port-3306" to "APP_KEY" or "DB_PORT_3306".
func FlagToEnvKey(flag string) (string, error) {
	if !strings.HasPrefix(flag, "--") {
		return "", fmt.Errorf("the flag must be in long format (--key), but received %s", flag)
	}
	clean := strings.TrimPrefix(flag, "--")
	envName := strings.ToUpper(strings.ReplaceAll(clean, "-", "_"))
	return envName, nil
}

// ResolveEnvFilePath resolves and validates the path to the .env file.
func ResolveEnvFilePath(rawPath string) (string, error) {
	if rawPath == "" {
		rawPath = ".env"
	}

	cleaned := filepath.Clean(rawPath)
	baseName := filepath.Base(cleaned)
	if baseName != ".env" && !strings.HasPrefix(baseName, ".env.") {
		return "", fmt.Errorf("the file should have .env prefix, but received %s", rawPath)
	}

	return cleaned, nil
}

// isAllowedWithoutConfirmation checks if an existing value is empty, null, true, or false.
func isAllowedWithoutConfirmation(val string) bool {
	v := strings.ToLower(strings.TrimSpace(val))
	return v == "" || v == "null" || v == "true" || v == "false"
}

// FindExampleFile checks if an example file (e.g. .env.example or <file>.example) exists.
func FindExampleFile(envFile string) string {
	// 1. Direct .example file: e.g. .env -> .env.example, .env.local -> .env.local.example
	candidate1 := envFile + ".example"
	if info, err := os.Stat(candidate1); err == nil && !info.IsDir() {
		return candidate1
	}

	// 2. Base .env.example in same directory if envFile is not .env
	dir := filepath.Dir(envFile)
	candidate2 := filepath.Join(dir, ".env.example")
	if candidate2 != candidate1 {
		if info, err := os.Stat(candidate2); err == nil && !info.IsDir() {
			return candidate2
		}
	}

	return ""
}

// askToUseExampleFile prompts the user whether they want to use an available example file.
func askToUseExampleFile(cmd *cobra.Command, reader *bufio.Reader, envFile, exampleFile string) (bool, error) {
	if yesFlag {
		return true, nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "The file '%s' does not exist. Would you like to use '%s'? (y/N): ", envFile, exampleFile)
	input, err := reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		return false, nil
	}
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	return input == "y" || input == "yes", nil
}

// LoadEnvFileWithFallback loads an env file, prompting to use an example file if the target file does not exist.
func LoadEnvFileWithFallback(cmd *cobra.Command, reader *bufio.Reader, envFile string) (map[string]string, string, bool, error) {
	var envMap map[string]string
	var envText string
	fileExists := true

	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		fileExists = false
		if exampleFile := FindExampleFile(envFile); exampleFile != "" {
			useExample, err := askToUseExampleFile(cmd, reader, envFile, exampleFile)
			if err != nil {
				return nil, "", false, err
			}
			if useExample {
				exampleMap, exampleText, err := gotenv.ReadEnvFile(exampleFile)
				if err != nil {
					return nil, "", false, fmt.Errorf("Unable to read %s file: %w", exampleFile, err)
				}
				envMap = exampleMap
				envText = exampleText
			}
		}
	} else if err != nil {
		return nil, "", false, fmt.Errorf("Unable to read %s file: %w", envFile, err)
	} else {
		loadedMap, loadedText, err := gotenv.ReadEnvFile(envFile)
		if err != nil {
			return nil, "", false, fmt.Errorf("Unable to read %s file: %w", envFile, err)
		}
		envMap = loadedMap
		envText = loadedText
	}

	if envMap == nil {
		envMap = make(map[string]string)
	}

	return envMap, envText, fileExists, nil
}

// askForConfirmation prompts the user to confirm overwriting a non-boolean/non-empty variable value.
func askForConfirmation(cmd *cobra.Command, reader *bufio.Reader, envName, currentVal, targetValue string) (bool, error) {
	if yesFlag {
		return true, nil
	}
	currentDisplay := currentVal
	if gotenv.IsSecret(envName, currentVal) {
		currentDisplay = gotenv.ObfuscateValue(currentVal)
	}
	targetDisplay := targetValue
	if gotenv.IsSecret(envName, targetValue) {
		targetDisplay = gotenv.ObfuscateValue(targetValue)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "The environment variable '%s' currently has value '%s'. Are you sure you want to change it to '%s'? (y/N): ", envName, currentDisplay, targetDisplay)
	input, err := reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		return false, nil
	}
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	return input == "y" || input == "yes", nil
}

// runBooleanCommand handles common logic for enable and disable commands.
func runBooleanCommand(cmd *cobra.Command, args []string, targetValue string) error {
	defer func() {
		yesFlag = false
	}()

	filePath := ".env"
	if len(args) > 0 {
		filePath = args[0]
	}

	envFile, err := ResolveEnvFilePath(filePath)
	if err != nil {
		return err
	}

	// Get unknown flags and convert them to env variables
	envFlags, err := GetUnknownFlags(cmd)
	if err != nil {
		return fmt.Errorf("Error reading command flags: %w", err)
	}

	if len(envFlags) == 0 {
		_ = cmd.Help()
		return nil
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	envMap, envText, fileExists, err := LoadEnvFileWithFallback(cmd, reader, envFile)
	if err != nil {
		return err
	}

	// Create a copy of the original envMap for comparison
	originalEnvMap := make(map[string]string, len(envMap))
	maps.Copy(originalEnvMap, envMap)

	flagKeys := make([]string, 0, len(envFlags))
	for flag := range envFlags {
		flagKeys = append(flagKeys, flag)
	}
	sort.Strings(flagKeys)

	changedEnvMap := make(map[string]string)

	for _, flag := range flagKeys {
		if !strings.HasPrefix(flag, "--") {
			return fmt.Errorf("The flag must be in long format, but received %s", flag)
		}
		envName, err := FlagToEnvKey(flag)
		if err != nil {
			return err
		}

		currentVal, exists := envMap[envName]
		if exists && !isAllowedWithoutConfirmation(currentVal) {
			confirmed, err := askForConfirmation(cmd, reader, envName, currentVal, targetValue)
			if err != nil {
				return err
			}
			if !confirmed {
				continue
			}
		}

		envMap[envName] = targetValue
		changedEnvMap[envName] = targetValue
	}

	if len(changedEnvMap) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No variables were updated.")
		return nil
	}

	// Write only if file didn't exist or if there are changes
	if !fileExists || !reflect.DeepEqual(originalEnvMap, envMap) || commentFlag != "" {
		commentsMap := make(map[string]string)
		if commentFlag != "" {
			for k := range changedEnvMap {
				commentsMap[k] = commentFlag
			}
		}
		if err := gotenv.WriteEnvFileWithComments(envFile, changedEnvMap, commentsMap, envText); err != nil {
			return fmt.Errorf("Error writing .env file: %w", err)
		}
	}
	commentFlag = ""

	// Sort before output
	rawResultKeys := make([]string, 0, len(changedEnvMap))
	for k := range changedEnvMap {
		rawResultKeys = append(rawResultKeys, k)
	}
	sort.Strings(rawResultKeys)

	preview := RenderSetPreview(envFile, rawResultKeys, changedEnvMap, nil)
	fmt.Fprint(cmd.OutOrStdout(), preview)
	return nil
}
