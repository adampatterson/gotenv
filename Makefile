.PHONY: docs build test

docs:
	go run ./internal/tools/docgen -out ./docs/cli

build:
	go build -o gotenv .

test:
	go test ./... -v
