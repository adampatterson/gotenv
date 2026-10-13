## gotenv unset

Blank the value of variables in a .env file

### Synopsis

Clear the value of one or more environment variables specified via flags from a .env file, keeping the variable (and its comments) in place.
If [file] is omitted, it defaults to .env in the current directory.
To remove the variable entirely, use "rm" or "delete" instead.

```
gotenv unset [file] [flags]
```

### Examples

```
  godot unset .env --app-key
  godot unset --statamic-license-key --statamic-pro-enabled
  godot unset .env.local --statamic-git-enabled
```

### Options

```
  -h, --help   help for unset
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

