.PHONY: docs

docs:
	go run ./internal/tools/docgen -out ./docs/cli
build:
	go build -o gotenv .
