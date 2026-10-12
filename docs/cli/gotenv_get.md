## gotenv get

Get the value of an environment variable from a .env file

### Synopsis

Retrieve the value of an environment variable specified via a long flag from a .env file.
The .env file should be named .env or start with .env. (e.g. .env.local).
If [file] is omitted, it defaults to .env in the current directory.

```
gotenv get [file] [flags]
```

### Examples

```
  godot get .env --app-key
  godot get --app-name
  godot get .env.local --statamic-license-key
  godot get --statamic-pro-enabled
  godot get .env --app-key --with-comments
  godot get .env --app-key -c
```

### Options

```
      --comments        Include comment when retrieving variable
  -h, --help            help for get
      --with-comment    Include comment when retrieving variable
  -c, --with-comments   Include comment when retrieving variable
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

