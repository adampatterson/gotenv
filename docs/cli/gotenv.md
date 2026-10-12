## gotenv

gotenv is a CLI tool for managing .env files

### Synopsis

gotenv is a command line utility to inspect, get, set, push (sync), and delete variables in .env files, inspired by DDEV's dotenv command.

```
gotenv [flags]
```

### Options

```
      --and strings           Additional target .env file(s) to update with the same values
  -c, --comment-text string   Comment to add above the environment variable
  -h, --help                  help for gotenv
      --with-comment string   Comment to add above the environment variable
  -y, --yes                   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv disable](gotenv_disable.md)	 - Disable environment variable(s) by setting their value to false
* [gotenv dotenv](gotenv_dotenv.md)	 - Commands for managing the contents of .env files
* [gotenv enable](gotenv_enable.md)	 - Enable environment variable(s) by setting their value to true
* [gotenv get](gotenv_get.md)	 - Get the value of an environment variable from a .env file
* [gotenv list](gotenv_list.md)	 - List all environment variables in a .env file
* [gotenv push](gotenv_push.md)	 - Push environment variable values from one .env file to another
* [gotenv rm](gotenv_rm.md)	 - Delete variables entirely from a .env file
* [gotenv set](gotenv_set.md)	 - Write values from the command line to a .env file
* [gotenv skill](gotenv_skill.md)	 - Generate a SKILL.md file for AI agent integrations
* [gotenv unset](gotenv_unset.md)	 - Blank the value of variables in a .env file

