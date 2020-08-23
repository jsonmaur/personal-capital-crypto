.PHONY: all format install
all: format install

format:
	@gofmt -s -w .

install:
	@go install ./...
