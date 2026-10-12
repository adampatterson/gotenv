# gotenv

A fast and friendly CLI tool written in Go using [Cobra](https://github.com/spf13/cobra) to inspect, get, set, push (sync), list, and delete environment variables in `.env` files, inspired by DDEV's `dotenv` command.

## Features

- **Dynamic Flag Mapping**: Pass arbitrary long flags (e.g. `--app-key=value` or `--statamic-license-key=secret`) which are automatically converted to uppercase environment variable keys (`APP_KEY="value"`, `STATAMIC_LICENSE_KEY="secret"`).
- **Structure & Comment Preservation**: Updates existing variables in place without wiping comments or disrupting existing formatting.
- **Safe Escaping**: Automatically escapes double quotes (`\"`) and shell dollar signs (`\$`) when writing values.
- **Multiple Environment Files**: Works with `.env`, `.env.local`, `.env.*` (e.g. `.ddev/.env.redis`).
- **DDEV-Compatible Subcommands**: Supports both `gotenv get ...` and `gotenv dotenv get ...` syntax.
- **List & JSON Export**: List variables formatted for shell sourcing (`--export`) or structured JSON (`--json`).

## Installation

### Using Homebrew (macOS / Linux)

```bash
brew install adampatterson/tap/gotenv
```

### From Source

```bash
# Clone the repository
git clone https://github.com/adampatterson/gotenv.git
cd gotenv

# Build binary
go build -o gotenv .

# Or install to your Go bin directory
go install
```

## Usage

### 1. Setting Variables (`set` / `add`)

Create or update variables in a `.env` file using flags:

```bash
# Sets APP_NAME and APP_URL in .env (default file)
gotenv set --app-name="Neon Nights" --app-url="http://statamic.test"

# Specify a specific .env file
gotenv set .env.local --statamic-license-key=secret123 --statamic-pro-enabled=true
gotenv set .ddev/.env.redis --redis-tag site-cache

# Update multiple files at once using --and
gotenv set .env --and="web/.env.local" --app-key=@generate
gotenv set cms/.env --and="web/.env.local" --and="api/.env" --shared-secret=@generate

# Automatically accept prompts using -y or --yes (e.g. initialize from .env.example if .env is missing)
gotenv set .env --app-name="New App" -y
```

**Adding Comments to Variables**: You can attach comments above variables when setting them using `comment="..."`, `-c`, `--comment-text`, or `--with-comment`:

```bash
# Using comment="..." prefix
gotenv comment="Production API Key" set .env --app-key=@generate

# Using -c flag
gotenv set .env --app-key=@generate -c "Production API Key"

# Setting/updating a comment on a variable without changing its existing value
gotenv --app-env -c "App env"
gotenv set .env --app-env -c "Application Environment"

# Setting an environment variable named COMMENT without conflict
gotenv set .env --comment="comment value"
gotenv set .env --comment="comment value" -c "Comment for COMMENT variable"

gotenv set .env --app-env -c "Adds a comment, but won't overwrite existing value"
```

**Generate Secrets Automatically**: You can use the magic values `@generate` or `@random` to automatically generate random secrets:
- For standard variables, a 16-byte random hex string is generated (equivalent to `$(openssl rand -hex 16)`):

```bash
gotenv set .env --statamic-license-key=@generate
```

- For Laravel application keys (`--app-key=@generate` or using `@key`, `@app-key`, `@base64`), a Laravel-compatible 32-byte base64-encoded key (`base64:...`) is generated:

```bash
gotenv set .env --app-key=@generate
```

### 2. Pushing / Synchronizing Variables (`push`)

Push environment variables from a source `.env` file (specified via `--from` or `--with`) to one or more target `.env` files (specified via `--to` and `--and`):

```bash
# Push specific variable from source to target
gotenv push --key=APP_KEY --from=web/.env.local --to=.env

# Push thing, from, to
gotenv push APP_KEY --from=web/.env.local --to=.env

# Push thing, from, to, and
gotenv push APP_KEY --from=web/.env.local --to=.env --and="api/.env"

# Push multiple keys with a custom comment
gotenv push DB_HOST DB_PORT --from=web/.env.local --to=.env -c "Synced from web"

# Push all variables from source file to target
gotenv push --from=web/.env.local --to=cms/.env --all
```

> **Path Best Practice**: All file paths are resolved relative to your current working directory (PWD). Both clean relative paths (e.g. `web/.env.local`) and `./` prefixes (e.g. `./web/.env.local`) are fully supported and normalized automatically.

### 3. Getting Variables (`get`)

Retrieve the raw value of a single variable:

```bash
# Get from default .env
gotenv get --app-name
# Output: Neon Nights

# Get from specific .env file
gotenv get .env.local --api-secret
# Output: secret123

# Include comments when retrieving variables using --with-comments (or -c, --comments)
gotenv get .env --app-key --with-comments
# Output:
# # Production API Key
# base64:...
```

### 4. Listing Variables (`list` / `show`)

List all key-value pairs defined in a file:

```bash
# Standard output
gotenv list .env

# Export format (ready to source in bash)
gotenv list .env --export

# JSON format
gotenv list .env --json
```

### 5. Enabling & Disabling Variables (`enable` / `disable`)

Set environment variables to `"true"` or `"false"`:

```bash
# Enable variables (sets to "true")
gotenv enable --app-debug
gotenv enable .env.local --statamic-pro-enabled --statamic-git-enabled

# Disable variables (sets to "false")
gotenv disable --app-debug
gotenv disable --statamic-stache-watcher
```

If the variable does not exist, it will be created. If the variable already exists with a value other than `empty`, `null`, `true`, or `false`, confirmation will be requested before overwriting (or pass `-y` / `--yes` to accept automatically).

### 6. Removing Variables (`unset` / `rm` / `delete`)

**Unsetting Values** (`unset`): Clears the value of one or more variables, keeping the variable (and any preceding comments) in the file:

```bash
gotenv unset --statamic-license-key
gotenv unset .env.local --statamic-pro-enabled --app-key
```

**Deleting Variables** (`rm` / `delete`): Removes one or more variables entirely, deleting the line from the file. Any preceding comments are always left untouched:

```bash
gotenv rm --statamic-license-key
gotenv delete .env.local --statamic-pro-enabled --app-key
```

### 7. Generating Agent Skills (`skill`)

Generate a `SKILL.md` file documenting `gotenv` CLI usage and conventions for AI coding agents and LLMs:

```bash
# Generate SKILL.md in the current directory
gotenv skill

# Generate SKILL.md into a specific directory
gotenv skill -out ~/PATH_NAME/
# or
gotenv skill --out ./docs/
```

### 8. DDEV Compatibility

You can also run all commands with the `dotenv` prefix matching DDEV CLI syntax:

```bash
gotenv dotenv get .env --app-key
gotenv dotenv set .env --statamic-pro-enabled=true
```

## Shell Completions

`gotenv` uses [Cobra](https://github.com/spf13/cobra), which automatically provides a `completion` subcommand for generating shell completion scripts for `bash`, `zsh`, `fish`, and `powershell` — no extra installation required beyond your shell's standard configuration.

**Bash**

```bash
gotenv completion bash > /etc/bash_completion.d/gotenv
# or, for user-local completions
gotenv completion bash > ~/.bash_completion
```

**Zsh**

```bash
gotenv completion zsh > "${fpath[1]}/_gotenv"
```

**Fish**

```bash
gotenv completion fish > ~/.config/fish/completions/gotenv.fish
```

**PowerShell**

```bash
gotenv completion powershell | Out-String | Invoke-Expression
```

For shell-specific setup details and caveats, run:

```bash
gotenv completion --help
```

## Running Tests

```bash
go test -v ./...
```

## Generating CLI Docs

Markdown reference docs for every command are generated into `docs/cli/` using [Cobra's doc generator](https://cobra.dev/docs/how-to-guides/clis-for-llms), making them easy for humans and LLMs to browse.

Regenerate them any time commands or flags change:

```bash
make docs
# or
go run ./internal/tools/docgen -out ./docs/cli
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on setting up the project, running tests, and opening pull requests.
