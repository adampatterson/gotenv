package cmd

import (
	"bufio"
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/adampatterson/gotenv/pkg/gotenv"

	"github.com/spf13/cobra"
)

var (
	pushFromFlag string
	pushToFlags  []string
	pushAndFlags []string
	pushKeyFlags []string
	pushAllFlag  bool
)

func isEnvPath(arg string) bool {
	cleaned := filepath.Clean(arg)
	base := filepath.Base(cleaned)
	return base == ".env" || strings.HasPrefix(base, ".env.")
}

func findMatchingKey(sourceMap map[string]string, key string) (string, bool) {
	if _, exists := sourceMap[key]; exists {
		return key, true
	}
	normalized := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
	if _, exists := sourceMap[normalized]; exists {
		return normalized, true
	}
	for srcKey := range sourceMap {
		if strings.EqualFold(srcKey, key) || strings.EqualFold(srcKey, normalized) {
			return srcKey, true
		}
	}
	return "", false
}

func runPushCommand(cmd *cobra.Command, args []string) error {
	defer func() {
		commentFlag = ""
		yesFlag = false
		pushFromFlag = ""
		pushToFlags = nil
		pushAndFlags = nil
		pushKeyFlags = nil
		pushAllFlag = false
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

	if pushFromFlag == "" {
		return fmt.Errorf("missing required source file flag --from (e.g. --from=\"web/.env.local\")")
	}

	sourceFile, err := ResolveEnvFilePath(pushFromFlag)
	if err != nil {
		return fmt.Errorf("Invalid source file: %w", err)
	}

	sourceMap, sourceText, err := gotenv.ReadEnvFile(sourceFile)
	if err != nil {
		return fmt.Errorf("Unable to read source file %s: %w", sourceFile, err)
	}

	var positionalKeys []string
	var targetFiles []string
	targetFiles = append(targetFiles, pushToFlags...)

	if len(pushToFlags) > 0 {
		// When --to is specified explicitly, all positional args are keys
		positionalKeys = append(positionalKeys, filteredArgs...)
	} else {
		// When --to is not specified, check if any positional arg is an env file path
		targetFound := false
		for _, arg := range filteredArgs {
			if !targetFound && isEnvPath(arg) {
				targetFiles = append(targetFiles, arg)
				targetFound = true
			} else {
				positionalKeys = append(positionalKeys, arg)
			}
		}
		if !targetFound {
			targetFiles = append(targetFiles, ".env")
		}
	}

	targetFiles = append(targetFiles, pushAndFlags...)

	var resolvedTargets []string
	seen := make(map[string]bool)
	for _, t := range targetFiles {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		resolved, err := ResolveEnvFilePath(t)
		if err != nil {
			return err
		}
		if resolved == sourceFile {
			return fmt.Errorf("source file and target file cannot be the same (%s)", resolved)
		}
		if !seen[resolved] {
			seen[resolved] = true
			resolvedTargets = append(resolvedTargets, resolved)
		}
	}

	envFlags, _, err := GetUnknownFlagsDetailed(cmd)
	if err != nil {
		return fmt.Errorf("Error reading command flags: %w", err)
	}

	if commentFlag == "" {
		if commentVal, hasComment := envFlags["--comment"]; hasComment {
			if len(envFlags) > 1 {
				commentFlag = commentVal
				delete(envFlags, "--comment")
			}
		}
	}

	var keysToPush []string
	seenKey := make(map[string]bool)

	addKey := func(keyInput string) error {
		matchedKey, found := findMatchingKey(sourceMap, keyInput)
		if !found {
			return fmt.Errorf("The environment variable '%s' not found in %s", keyInput, sourceFile)
		}
		if !seenKey[matchedKey] {
			seenKey[matchedKey] = true
			keysToPush = append(keysToPush, matchedKey)
		}
		return nil
	}

	if pushAllFlag {
		for k := range sourceMap {
			keysToPush = append(keysToPush, k)
		}
		sort.Strings(keysToPush)
	} else {
		// Keys from --key / -k flags
		for _, kf := range pushKeyFlags {
			// Support comma-separated keys if provided
			parts := strings.Split(kf, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" {
					if err := addKey(part); err != nil {
						return err
					}
				}
			}
		}

		// Keys from positional arguments
		for _, pk := range positionalKeys {
			pk = strings.TrimSpace(pk)
			if pk != "" {
				if err := addKey(pk); err != nil {
					return err
				}
			}
		}

		// Keys from dynamic flags (e.g. --app-secret, --db-host)
		for flag := range envFlags {
			if !strings.HasPrefix(flag, "--") {
				return fmt.Errorf("The flag must be in long format, but received %s", flag)
			}
			envName, err := FlagToEnvKey(flag)
			if err != nil {
				return err
			}
			if err := addKey(envName); err != nil {
				return err
			}
		}
		sort.Strings(keysToPush)
	}

	if len(keysToPush) == 0 {
		_ = cmd.Help()
		return nil
	}

	reader := bufio.NewReader(cmd.InOrStdin())

	for _, targetFile := range resolvedTargets {
		targetMap, targetText, fileExists, err := LoadEnvFileWithFallback(cmd, reader, targetFile)
		if err != nil {
			return err
		}

		originalTargetMap := make(map[string]string, len(targetMap))
		maps.Copy(originalTargetMap, targetMap)

		changedEnvMap := make(map[string]string)
		commentsMap := make(map[string]string)

		for _, key := range keysToPush {
			val := sourceMap[key]
			targetMap[key] = val
			changedEnvMap[key] = val

			if commentFlag != "" {
				commentsMap[key] = commentFlag
			} else {
				srcComment := gotenv.GetEnvVarComment(sourceText, key)
				if srcComment != "" {
					commentsMap[key] = srcComment
				}
			}
		}

		if !fileExists || !reflect.DeepEqual(originalTargetMap, targetMap) || commentFlag != "" {
			if err := gotenv.WriteEnvFileWithComments(targetFile, changedEnvMap, commentsMap, targetText); err != nil {
				return fmt.Errorf("Error writing %s file: %w", targetFile, err)
			}
		}

		rawResultKeys := make([]string, 0, len(changedEnvMap))
		for k := range changedEnvMap {
			rawResultKeys = append(rawResultKeys, k)
		}
		sort.Strings(rawResultKeys)

		preview := RenderSetPreview(targetFile, rawResultKeys, changedEnvMap, commentsMap)
		fmt.Fprint(cmd.OutOrStdout(), preview)
	}

	return nil
}

var pushCmd = &cobra.Command{
	Use:     "push [key...] [file]",
	Aliases: []string{"sync", "synchronize"},
	Short:   "Push environment variable values from one .env file to another",
	Long: `Push environment variable values from a source .env file to one or more target .env files.
The source file is specified using --from (or --with).
The target file can be specified using --to (defaults to .env) and additional target files via --and.
Variables to push can be specified using --key (or -k), as positional arguments, or via dynamic flags (e.g. --app-key).
To push all variables present in the source file, use the --all (or -a) flag.

Paths are evaluated relative to the current working directory (PWD).`,
	Example: `  godot push --key=APP_KEY --from=web/.env.local --to=.env
  godot push APP_KEY --from=web/.env.local --to=.env
  godot push APP_KEY --from=web/.env.local --to=.env --and=api/.env
  godot push --from=web/.env.local --to=.env --all
  godot push .env --from=web/.env.local --app-key
  godot push --from=web/.env.local --app-secret`,
	Args: cobra.ArbitraryArgs,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPushCommand(cmd, args)
	},
}

func init() {
	pushCmd.Flags().StringVar(&pushFromFlag, "from", "", "Source .env file to copy values from")
	pushCmd.Flags().StringVar(&pushFromFlag, "with", "", "Source .env file to copy values from (alias for --from)")
	pushCmd.Flags().StringSliceVar(&pushToFlags, "to", nil, "Target .env file(s) to push values to")
	pushCmd.Flags().StringSliceVar(&pushAndFlags, "and", nil, "Additional target .env file(s) to push values to")
	pushCmd.Flags().StringSliceVarP(&pushKeyFlags, "key", "k", nil, "Key(s) to push from the source file")
	pushCmd.Flags().BoolVarP(&pushAllFlag, "all", "a", false, "Push all environment variables from source file")
	pushCmd.Flags().StringVarP(&commentFlag, "comment-text", "c", "", "Comment to add above the pushed environment variable")
	pushCmd.Flags().StringVar(&commentFlag, "with-comment", "", "Comment to add above the pushed environment variable")
}
