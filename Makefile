# Settings
# Disable makefile default values and rules
# Unconditionally make all targets.
MAKEFLAGS=--no-builtin-rules --no-builtin-variables --always-make

# Rules
.DEFAULT_GOAL := gen

fmt:
	go fmt ./...
	go tool modernize -fix ./...

lint:
	go vet ./...

test:
	go test ./...

benchmark:
	go test -bench=. -run=^$ ./...
