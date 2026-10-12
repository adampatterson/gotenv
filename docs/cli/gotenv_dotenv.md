## gotenv dotenv

Commands for managing the contents of .env files

### Synopsis

Subcommands for inspecting, getting, and setting values in .env files (compatible with DDEV dotenv syntax).

```
gotenv dotenv [command] [flags]
```

### Examples

```
  godot dotenv get --app-key
  godot dotenv set --app-name="Statamic Peak"
  godot dotenv unset --app-key
  godot dotenv list
```

### Options

```
  -c, --comment-text string   Comment to add above the environment variable
  -h, --help                  help for dotenv
      --with-comment string   Comment to add above the environment variable
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files
* [gotenv dotenv disable](gotenv_dotenv_disable.md)	 - Disable environment variable(s) by setting their value to false
* [gotenv dotenv enable](gotenv_dotenv_enable.md)	 - Enable environment variable(s) by setting their value to true
* [gotenv dotenv get](gotenv_dotenv_get.md)	 - Get the value of an environment variable from a .env file
* [gotenv dotenv list](gotenv_dotenv_list.md)	 - List all environment variables in a .env file
* [gotenv dotenv rm](gotenv_dotenv_rm.md)	 - Delete variables entirely from a .env file
* [gotenv dotenv set](gotenv_dotenv_set.md)	 - Write values from the command line to a .env file
* [gotenv dotenv skill](gotenv_dotenv_skill.md)	 - Generate a SKILL.md file for AI agent integrations
* [gotenv dotenv unset](gotenv_dotenv_unset.md)	 - Blank the value of variables in a .env file

