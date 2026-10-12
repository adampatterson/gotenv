package main

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/adampatterson/gotenv/cmd"
)

func main() {
	out := flag.String("out", "./docs/cli", "output directory for generated docs")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	root := cmd.GetRootCmd()
	root.DisableAutoGenTag = true

	if err := doc.GenMarkdownTree(root, *out); err != nil {
		log.Fatal(err)
	}
}
