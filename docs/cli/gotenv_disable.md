## gotenv disable

Disable environment variable(s) by setting their value to false

### Synopsis

Set one or more environment variables to "false" in a .env file.
The .env file should be named .env or start with .env. (e.g. .env.local).
If [file] is omitted, it defaults to .env in the current directory.

If the variable does not exist, it will be created.
If the variable already exists with a value other than empty, null, true, or false, confirmation will be requested before making the change.

```
gotenv disable [file] [flags]
```

### Examples

```
  godot disable --app-debug
  godot disable .env --statamic-stache-watcher
  godot disable .env.local --statamic-git-enabled --statamic-pro-enabled
```

### Options

```
  -h, --help   help for disable
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

