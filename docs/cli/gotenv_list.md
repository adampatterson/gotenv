## gotenv list

List all environment variables in a .env file

### Synopsis

Display all variables defined in a .env file.
If [file] is omitted, it defaults to .env in the current directory.

```
gotenv list [file] [flags]
```

### Examples

```
  godot list
  godot list .env.local
  godot list --json
  godot list --export
```

### Options

```
  -e, --export   Output with export prefix (e.g. export KEY="value")
  -h, --help     help for list
  -j, --json     Output list formatted as JSON
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

