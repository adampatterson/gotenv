## gotenv run

Run a command with the variables from .env files in its environment

### Synopsis

Load one or more .env files and run a command with those variables added to its environment.
Values from the files override variables that are already set. If --env-file is given more than once,
later files override earlier ones. If --env-file is omitted, it defaults to .env in the current directory.
The files are only read. The command inherits stdin, stdout, stderr, and signals, and gotenv exits with its exit code.

```
gotenv run [--env-file=file]... [--] command [args...] [flags]
```

### Examples

```
  godot run -- ddev k6 run
  godot run --env-file=.env.local -- npm test
  godot run --env-file=.env --env-file=.env.local -- printenv APP_KEY
```

### Options

```
      --env-file stringArray   Path to a .env file to load (repeatable, defaults to .env)
  -h, --help                   help for run
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

