package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"maps"
	"reflect"
	"sort"
	"strings"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

func generateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateAppKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "base64:" + base64.StdEncoding.EncodeToString(bytes), nil
}

func isAppKeyValue(envName, value string) bool {
	lowerVal := strings.ToLower(value)
	if envName == "APP_KEY" {
		return lowerVal == "@generate" || lowerVal == "@random" || lowerVal == "@key" || lowerVal == "@app-key" || lowerVal == "@laravel-key" || lowerVal == "@base64" || strings.HasPrefix(lowerVal, "@generate:") || strings.HasPrefix(lowerVal, "@random:")
	}
	return lowerVal == "@key" || lowerVal == "@app-key" || lowerVal == "@laravel-key" || lowerVal == "@base64" || lowerVal == "@generate:key" || lowerVal == "@generate:app-key" || lowerVal == "@generate:base64" || lowerVal == "@random:key" || lowerVal == "@random:app-key" || lowerVal == "@random:base64"
}

func isRandomHexValue(value string) bool {
	lowerVal := strings.ToLower(value)
	return lowerVal == "@generate" || lowerVal == "@random" || lowerVal == "@hex" || lowerVal == "@generate:hex" || lowerVal == "@random:hex"
}

var (
	commentFlag string
	andFlags    []string
)

func runSetCommand(cmd *cobra.Command, args []string) error {
	defer func() {
		commentFlag = ""
		yesFlag = false
		andFlags = nil
	}()

	var filteredArgs []string
	for _, arg := range args {
		lower := strings.ToLower(arg)
		if strings.HasPrefix(lower, "comment=") {
			val := arg[8:]
			commentFlag = strings.Trim(val, `"'`)
		} else if strings.HasPrefix(lower, "comment:") {
			val := arg[8:]
			commentFlag = strings.Trim(val, `"'`)
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) > 1 {
		return fmt.Errorf("accepts at most 1 arg(s), received %d", len(filteredArgs))
	}

	filePath := ".env"
	if len(filteredArgs) > 0 {
		filePath = filteredArgs[0]
	}

	targetFiles := []string{filePath}
	for _, extra := range andFlags {
		extra = strings.TrimSpace(extra)
		if extra != "" {
			targetFiles = append(targetFiles, extra)
		}
	}

	var resolvedFiles []string
	seen := make(map[string]bool)
	for _, f := range targetFiles {
		resolved, err := ResolveEnvFilePath(f)
		if err != nil {
			return err
		}
		if !seen[resolved] {
			seen[resolved] = true
			resolvedFiles = append(resolvedFiles, resolved)
		}
	}

	// Get unknown flags and convert them to env variables
	envFlags, hasExplicitVal, err := GetUnknownFlagsDetailed(cmd)
	if err != nil {
		return fmt.Errorf("Error reading command flags: %w", err)
	}

	// Disambiguate --comment flag:
	// If commentFlag is not set and --comment is present:
	// If other env flags are present (len > 1), --comment is the comment text for the other flag(s).
	// If only --comment is present (len == 1), the user wants to set COMMENT="value".
	if commentFlag == "" {
		if commentVal, hasComment := envFlags["--comment"]; hasComment {
			if len(envFlags) > 1 {
				commentFlag = commentVal
				delete(envFlags, "--comment")
				delete(hasExplicitVal, "--comment")
			}
		}
	}

	if len(envFlags) == 0 {
		_ = cmd.Help()
		return nil
	}

	// Pre-calculate generated dynamic values (such as @generate or @random)
	// so that if multiple files are updated via --and, all files receive the same generated value.
	evaluatedValues := make(map[string]string)
	explicitFlags := make(map[string]bool)
	hasUnknownFlags := false

	for flag, value := range envFlags {
		if !strings.HasPrefix(flag, "--") {
			return fmt.Errorf("The flag must be in long format, but received %s", flag)
		}
		envName, err := FlagToEnvKey(flag)
		if err != nil {
			return err
		}

		if hasExplicitVal[flag] {
			if isAppKeyValue(envName, value) {
				keyValue, err := generateAppKey()
				if err != nil {
					return fmt.Errorf("Failed to generate app key for %s: %w", envName, err)
				}
				value = keyValue
			} else if isRandomHexValue(value) {
				hexValue, err := generateRandomHex(16)
				if err != nil {
					return fmt.Errorf("Failed to generate random hex for %s: %w", envName, err)
				}
				value = hexValue
			}
			evaluatedValues[envName] = value
			explicitFlags[envName] = true
			hasUnknownFlags = true
		} else {
			explicitFlags[envName] = false
			hasUnknownFlags = true
		}
	}

	if !hasUnknownFlags {
		_ = cmd.Help()
		return nil
	}

	reader := bufio.NewReader(cmd.InOrStdin())

	for _, envFile := range resolvedFiles {
		envMap, envText, fileExists, err := LoadEnvFileWithFallback(cmd, reader, envFile)
		if err != nil {
			return err
		}

		// Create a copy of the original envMap for comparison
		originalEnvMap := make(map[string]string, len(envMap))
		maps.Copy(originalEnvMap, envMap)

		changedEnvMap := make(map[string]string)

		for envName, isExplicit := range explicitFlags {
			if isExplicit {
				val := evaluatedValues[envName]
				envMap[envName] = val
				changedEnvMap[envName] = val
			} else {
				// Flag provided without explicit value (e.g. --app-env -c "Comment" or just --app-env)
				if existingVal, exists := envMap[envName]; exists {
					changedEnvMap[envName] = existingVal
				} else {
					envMap[envName] = ""
					changedEnvMap[envName] = ""
				}
			}
		}

		commentsMap := make(map[string]string)
		if commentFlag != "" {
			for k := range changedEnvMap {
				commentsMap[k] = commentFlag
			}
		}

		// Write only if file didn't exist or if there are changes
		if !fileExists || !reflect.DeepEqual(originalEnvMap, envMap) || commentFlag != "" {
			if err := gotenv.WriteEnvFileWithComments(envFile, changedEnvMap, commentsMap, envText); err != nil {
				return fmt.Errorf("Error writing %s file: %w", envFile, err)
			}
		}

		// Sort before output
		rawResultKeys := make([]string, 0, len(changedEnvMap))
		for k := range changedEnvMap {
			rawResultKeys = append(rawResultKeys, k)
		}
		sort.Strings(rawResultKeys)

		preview := RenderSetPreview(envFile, rawResultKeys, changedEnvMap, commentsMap)
		fmt.Fprint(cmd.OutOrStdout(), preview)
	}

	return nil
}

var setCmd = &cobra.Command{
	Use:     "set [file]",
	Aliases: []string{"add"},
	Short:   "Write values from the command line to a .env file",
	Long: `Create or update a .env file with values specified via long flags from the command line.
Flags in the format --env-key=value will be converted to environment variable names
like ENV_KEY="value". The .env file should be named .env or start with .env. (e.g. .env.local).
If [file] is omitted, it defaults to .env in the current directory.

Paths are evaluated relative to the current working directory (PWD).
You can specify additional files to update simultaneously using the --and flag.

You can add comments to variables using comment="Comment text", -c "Comment text",
--comment-text="Comment text", or --with-comment="Comment text".

You can use the magic values "@generate" or "@random" to automatically generate
a 16-byte random hex string (equivalent to openssl rand -hex 16).
For Laravel application keys (e.g. APP_KEY or magic values "@key", "@app-key", "@base64"),
a 32-byte base64-encoded key (base64:...) will be generated.`,
	Example: `  godot set .env --app-name="Statamic Peak"
  godot set --app-url=http://statamic.test --app-debug=true
  godot set .env --statamic-license-key=@generate
  godot set .env.local --statamic-pro-enabled=true --statamic-git-enabled=false
  godot set .env --and="web/.env.local" --app-secret=@generate
  godot set cms/.env --and="web/.env.local" --app-name="Statamic Peak"
  godot set --app-key=@generate
  godot comment="Production API Key" set .env --app-key=@generate
  godot set .env --app-key=@generate -c "Production API Key"`,
	Args: cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSetCommand(cmd, args)
	},
}

func init() {
	setCmd.Flags().StringVarP(&commentFlag, "comment-text", "c", "", "Comment to add above the environment variable")
	setCmd.Flags().StringVar(&commentFlag, "with-comment", "", "Comment to add above the environment variable")
	setCmd.Flags().StringSliceVar(&andFlags, "and", nil, "Additional target .env file(s) to update with the same values")
}
