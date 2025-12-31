# Settings
# Disable makefile default values and rules
# Unconditionally make all targets.
MAKEFLAGS=--no-builtin-rules --no-builtin-variables --always-make

# Rules
.DEFAULT_GOAL := gen

tidy:
	go mod tidy

fmt:
	go fmt ./...
	go tool modernize -fix ./...

lint:
	go vet ./...

test:
	go test ./...

checks: tidy fmt lint test

benchmark:
	go test -bench=. -run=^$ ./...
