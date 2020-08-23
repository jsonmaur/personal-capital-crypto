.PHONY: all format build
all: format build

format:
	@gofmt -s -w .

build:
	@go build -o dist/macos ./...
