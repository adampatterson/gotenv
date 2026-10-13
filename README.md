# gotenv

A Go CLI tool to inspect, set, sync, list, and delete environment variables in `.env` files.

## Features

- **Dynamic Flag Mapping**: Pass arbitrary long flags (e.g. `--app-key=value` or `--statamic-license-key=secret`) which convert to uppercase environment variable keys (`APP_KEY="value"`, `STATAMIC_LICENSE_KEY="secret"`).
- **Structure & Comment Preservation**: Updates existing variables in place without wiping comments or disrupting existing formatting.
- **Safe Escaping**: Escapes double quotes (`\"`) and shell dollar signs (`\$`) when writing values.
- **Multiple Environment Files**: Works with `.env`, `.env.local`, `.env.*` (e.g. `.ddev/.env.redis`).
- **DDEV-Compatible Subcommands**: Supports both `gotenv get ...` and `gotenv dotenv get ...` syntax.
- **List & JSON Export**: List variables formatted for shell sourcing (`--export`) or structured JSON (`--json`).

## Installation

### Using Homebrew (macOS / Linux)

```bash
brew install adampatterson/tap/gotenv
```

### Pre-built Binaries

Download pre-compiled binaries for macOS, Linux, and Windows from [GitHub Releases](https://github.com/adampatterson/gotenv/releases).

#### macOS Security Notice

If macOS blocks the binary from opening, use one of these options:

##### Option 1: Terminal Command

Remove quarantine attributes:

```bash
chmod +x gotenv
xattr -cr gotenv
```

##### Option 2: macOS System Settings

1. Run `./gotenv` in your terminal to trigger the system prompt.
2. Open **System Settings** > **Privacy & Security**.
3. Scroll to the **Security** section.
4. Click **Open Anyway** next to the `gotenv` notice and confirm.

### From Source

```bash
# Clone the repository
git clone https://github.com/adampatterson/gotenv.git
cd gotenv

# Build binary
make build

# Or install to your Go bin directory
go install
```

## Usage

### Quick Start

The three most common operations:

```bash
# 1. Set a variable
gotenv set --app-name="Neon Nights"

# 2. Get a variable value
gotenv get --app-name

# 3. Push a variable to another .env file
gotenv push APP_KEY --from=.env --to=.env.local
```

### Mental Model: Data Flow

`gotenv` uses two data flow patterns:

- **Single-file commands (`set`, `get`, `rm`, `enable`, `disable`)**:
  `action -> target_file -> variables`
  The positional file argument is the file that you read or change.
  ```bash
  gotenv set .env --app-key=secret
  gotenv get .env --app-key
  ```

- **Transfer commands (`push`)**:
  `action -> source_file -> target_file -> variables`
  Data moves from a source file to one or more target files.
  ```bash
  gotenv push .env --to=web/.env.local --app-key
  gotenv push APP_KEY --from=.env --to=web/.env.local
  ```

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

**Generate Secrets Automatically**: You can use the magic values `@generate` or `@random` to generate random secrets:
- For standard variables, the tool generates a 16-byte random hex string (equivalent to `$(openssl rand -hex 16)`):

```bash
gotenv set .env --statamic-license-key=@generate
```

- For Laravel application keys (`--app-key=@generate` or using `@key`, `@app-key`, `@base64`), the tool generates a 32-byte base64-encoded key for Laravel (`base64:...`):

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

# Monorepo sync (e.g. pushing tokens across sibling packages)
gotenv push .env --to=../web/.env.local --refresh-token
gotenv push REFRESH_TOKEN --from=.env --to=../web/.env.local
gotenv push ../web/.env.local --from=.env --refresh-token
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

If the variable does not exist, the command creates it. If the variable already exists with a value other than `empty`, `null`, `true`, or `false`, the command prompts for confirmation before it overwrites values (or pass `-y` / `--yes` to accept automatically).

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

`gotenv` uses [Cobra](https://github.com/spf13/cobra), which includes a `completion` subcommand for `bash`, `zsh`, `fish`, and `powershell`. No extra installation is required beyond standard shell configuration.

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
make test
# or
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

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for details on setting up the project, running tests, and opening pull requests.

## Note on Raw Output & Security

`gotenv` is designed for managing and inspecting local environment files (`.env`), operating directly on local files that the developer already has filesystem permissions to read. Like standard POSIX utilities (`env`, `printenv`) and other dotenv CLI tools, `gotenv` outputs raw values verbatim without masking to prioritize operational transparency, scriptability (e.g. `--export`, shell evaluation), and straightforward local debugging. 

Unlike dedicated secret management platforms (such as HashiCorp Vault, 1Password CLI, or AWS Secrets Manager) which mask or redact secret values by default, `gotenv` treats all `.env` entries uniformly. Users should exercise standard caution when sharing screens, recording demos, or piping command output in shared or CI environments.
