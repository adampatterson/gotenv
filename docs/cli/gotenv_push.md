## gotenv push

Push environment variable values from one .env file to another

### Synopsis

Push environment variable values from a source .env file to one or more target .env files.
The source file is specified using --from (or --with).
The target file can be specified using --to (defaults to .env) and additional target files via --and.
Variables to push can be specified using --key (or -k), as positional arguments, or via dynamic flags (e.g. --app-key).
To push all variables present in the source file, use the --all (or -a) flag.

Paths are evaluated relative to the current working directory (PWD).

```
gotenv push [key...] [file] [flags]
```

### Examples

```
  godot push --key=APP_KEY --from=web/.env.local --to=.env
  godot push APP_KEY --from=web/.env.local --to=.env
  godot push APP_KEY --from=web/.env.local --to=.env --and=api/.env
  godot push --from=web/.env.local --to=.env --all
  godot push .env --from=web/.env.local --app-key
  godot push --from=web/.env.local --app-secret
```

### Options

```
  -a, --all                   Push all environment variables from source file
      --and strings           Additional target .env file(s) to push values to
  -c, --comment-text string   Comment to add above the pushed environment variable
      --from string           Source .env file to copy values from
  -h, --help                  help for push
  -k, --key strings           Key(s) to push from the source file
      --to strings            Target .env file(s) to push values to
      --with string           Source .env file to copy values from (alias for --from)
      --with-comment string   Comment to add above the pushed environment variable
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

