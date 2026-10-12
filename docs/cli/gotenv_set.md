## gotenv set

Write values from the command line to a .env file

### Synopsis

Create or update a .env file with values specified via long flags from the command line.
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
a 32-byte base64-encoded key (base64:...) will be generated.

```
gotenv set [file] [flags]
```

### Examples

```
  godot set .env --app-name="Statamic Peak"
  godot set --app-url=http://statamic.test --app-debug=true
  godot set .env --statamic-license-key=@generate
  godot set .env.local --statamic-pro-enabled=true --statamic-git-enabled=false
  godot set .env --and="web/.env.local" --app-secret=@generate
  godot set cms/.env --and="web/.env.local" --app-name="Statamic Peak"
  godot set --app-key=@generate
  godot comment="Production API Key" set .env --app-key=@generate
  godot set .env --app-key=@generate -c "Production API Key"
```

### Options

```
      --and strings           Additional target .env file(s) to update with the same values
  -c, --comment-text string   Comment to add above the environment variable
  -h, --help                  help for set
      --with-comment string   Comment to add above the environment variable
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

