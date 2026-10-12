## gotenv skill

Generate a SKILL.md file for AI agent integrations

### Synopsis

Generate a SKILL.md file documenting gotenv CLI usage and capabilities for AI coding agents and LLMs.
If [path] or --out is omitted, it defaults to SKILL.md in the current directory.
If a directory is specified (e.g. ~/my-skills/ or ./skills), SKILL.md will be created inside that directory.

```
gotenv skill [path] [flags]
```

### Examples

```
  godot skill
  godot skill -out ~/skills/
  godot skill --out ./skills/gotenv/
  godot skill ./docs/SKILL.md
```

### Options

```
  -h, --help            help for skill
  -o, --out string      Output directory or file path for SKILL.md
      --output string   Output directory or file path for SKILL.md
```

### Options inherited from parent commands

```
  -y, --yes   Automatically answer yes to prompts (e.g. using example file or confirming changes)
```

### SEE ALSO

* [gotenv](gotenv.md)	 - gotenv is a CLI tool for managing .env files

