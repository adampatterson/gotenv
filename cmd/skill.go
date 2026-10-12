package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var skillOutFlag string

// SkillDocContent is the Markdown content written to SKILL.md for AI agent integrations.
const SkillDocContent = `---
name: gotenv
description: Inspect, get, set, push (sync), unset, enable, disable, and delete environment variables in .env files using the gotenv CLI tool.
---

# gotenv Skill

` + "```" + `
gotenv [command] [file] [flags]
` + "```" + `

` + "`gotenv`" + ` (and its alias ` + "`godot`" + `) is a fast, ergonomic CLI utility for managing, inspecting, and synchronizing ` + "`.env`" + ` files. It preserves formatting and comments while supporting type-safe operations, key generation, and boolean toggles.

## When to Use

Use ` + "`gotenv`" + ` when you need to:
- Read environment variable values from ` + "`.env`" + ` files.
- Set or update environment variables while keeping surrounding comments and structure intact.
- Generate secure random keys (` + "`APP_KEY`" + `, hex tokens, base64 hashes).
- Toggle boolean flags (` + "`true`" + ` / ` + "`false`" + `) safely.
- List or export variables in plain, bash ` + "`export`" + `, or JSON format.
- Push / sync variables between different environment files (e.g. ` + "`.env.example`" + ` to ` + "`.env`" + `).
- Blank out (` + "`unset`" + `) or completely remove (` + "`rm`" + ` / ` + "`delete`" + `) environment variables.

---

## Command Reference

If ` + "`[file]`" + ` is omitted in any command, it defaults to ` + "`.env`" + ` in the current directory.

### 1. Setting Variables (` + "`set`" + ` or root command)

Set one or more variables using flag syntax (` + "`--key=value`" + ` or ` + "`--KEY=VALUE`" + `). Hyphens are automatically converted to uppercase underscores (e.g., ` + "`--app-name`" + ` -> ` + "`APP_NAME`" + `).

` + "```bash" + `
# Set variables explicitly
gotenv set .env --app-name="My Application" --app-port=8080

# Direct command execution without 'set'
gotenv .env --app-env=production --database-url="postgres://user:pass@localhost:5432/db"

# Update default .env
gotenv --app-key=secret

# Add a comment above the variable
gotenv set .env --api-key="xyz123" -c "Production API key"
gotenv set .env --api-key="xyz123" --comment-text="Production API key"

# Update multiple target files simultaneously
gotenv set .env --app-env=staging --and .env.staging --and .env.test
` + "```" + `

#### Dynamic Value Generators
- ` + "`@generate`" + `, ` + "`@key`" + `, ` + "`@base64`" + `: Generates a 32-byte base64 key (` + "`base64:...`" + `) for ` + "`APP_KEY`" + `.
- ` + "`@hex`" + `, ` + "`@generate:hex`" + `: Generates a 32-character random hex string.

` + "```bash" + `
gotenv set .env --app-key=@generate
gotenv set .env --auth-token=@hex
` + "```" + `

---

### 2. Getting Variables (` + "`get`" + `)

Retrieve the value of an environment variable.

` + "```bash" + `
# Retrieve variable value
gotenv get .env --app-key
gotenv get --app-name

# Retrieve variable with comment table
gotenv get .env --app-key --with-comments
gotenv get .env --app-key -c
` + "```" + `

---

### 3. Listing Variables (` + "`list`" + ` / ` + "`show`" + `)

List all variables defined in a ` + "`.env`" + ` file.

` + "```bash" + `
# Standard list view
gotenv list .env

# Bash export format (sourceable)
gotenv list .env --export

# JSON format
gotenv list .env --json
` + "```" + `

---

### 4. Toggling Boolean Values (` + "`enable`" + ` / ` + "`disable`" + `)

Quickly toggle boolean variables to ` + `"true"` + ` or ` + `"false"` + `.

` + "```bash" + `
# Enable (sets value to "true")
gotenv enable .env --app-debug --statamic-pro-enabled

# Disable (sets value to "false")
gotenv disable .env --app-debug
` + "```" + `

---

### 5. Unsetting & Deleting Variables (` + "`unset`" + `, ` + "`rm`" + ` / ` + "`delete`" + `)

- ` + "`unset`" + `: Clears the variable's value, keeping the key and preceding comments in the file.
- ` + "`rm`" + ` / ` + "`delete`" + `: Removes the variable line entirely, preserving preceding comments.

` + "```bash" + `
# Blank value (keeps KEY= in file)
gotenv unset .env --statamic-license-key

# Remove variable line entirely
gotenv rm .env --statamic-license-key
gotenv delete .env.local --debug --telemetry-enabled
` + "```" + `

---

### 6. Pushing & Syncing Variables (` + "`push`" + `)

Copy variable values from a source file to one or more target files.

` + "```bash" + `
# Push specific variables from source to target
gotenv push --from .env.example .env --app-name --app-url

# Push all variables from source to target
gotenv push --from .env.example .env --all

# Push to multiple target files
gotenv push --from .env --to .env.local --and .env.staging --all
` + "```" + `

---

### 7. Non-Interactive / Automation Mode (` + "`-y`" + ` / ` + "`--yes`" + `)

Automatically accept prompts when creating missing files or overwriting non-standard values:

` + "```bash" + `
gotenv set .env --app-debug=true -y
gotenv enable .env --app-debug --yes
` + "```" + `

---

### 8. Generating Skills & Documentation (` + "`skill`" + `)

Generate a ` + "`SKILL.md`" + ` file for AI agents:

` + "```bash" + `
gotenv skill
gotenv skill -out ~/skills/gotenv/
gotenv skill --out ./docs/
` + "```" + `
`

func expandHomePath(p string) (string, error) {
	if p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

var skillCmd = &cobra.Command{
	Use:   "skill [path]",
	Short: "Generate a SKILL.md file for AI agent integrations",
	Long: `Generate a SKILL.md file documenting gotenv CLI usage and capabilities for AI coding agents and LLMs.
If [path] or --out is omitted, it defaults to SKILL.md in the current directory.
If a directory is specified (e.g. ~/my-skills/ or ./skills), SKILL.md will be created inside that directory.`,
	Example: `  godot skill
  godot skill -out ~/skills/
  godot skill --out ./skills/gotenv/
  godot skill ./docs/SKILL.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		defer func() {
			skillOutFlag = ""
		}()

		target := skillOutFlag
		if target == "" && len(args) > 0 {
			target = args[0]
		}
		if target == "" {
			target = "SKILL.md"
		}

		expandedPath, err := expandHomePath(target)
		if err != nil {
			return fmt.Errorf("error expanding path %s: %w", target, err)
		}

		var destFile string
		var destDir string

		fi, statErr := os.Stat(expandedPath)
		if statErr == nil && fi.IsDir() {
			destDir = expandedPath
			destFile = filepath.Join(expandedPath, "SKILL.md")
		} else if strings.HasSuffix(target, "/") || strings.HasSuffix(target, "\\") || filepath.Ext(expandedPath) != ".md" {
			destDir = expandedPath
			destFile = filepath.Join(expandedPath, "SKILL.md")
		} else {
			destFile = expandedPath
			destDir = filepath.Dir(expandedPath)
		}

		if destDir != "" && destDir != "." {
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destDir, err)
			}
		}

		if err := os.WriteFile(destFile, []byte(SkillDocContent), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", destFile, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Successfully generated %s\n", destFile)
		return nil
	},
}

func init() {
	skillCmd.Flags().StringVarP(&skillOutFlag, "out", "o", "", "Output directory or file path for SKILL.md")
	skillCmd.Flags().StringVar(&skillOutFlag, "output", "", "Output directory or file path for SKILL.md")
}
