## gotenv rm

Delete variables entirely from a .env file

### Synopsis

Remove one or more environment variables specified via flags from a .env file, deleting the variable line entirely.
If [file] is omitted, it defaults to .env in the current directory.
To keep the variable but blank its value instead, use "unset".

```
gotenv rm [file] [flags]
```

### Examples

```
  godot rm .env --app-key
  godot delete --statamic-license-key --statamic-pro-enabled
  godot rm .env.local --statamic-git-enabled
```

### Options

```
  -h, --help   help for rm
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files
