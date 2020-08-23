.PHONY: all format install
all: format install

format:
	@gofmt -s -w .

install:
	@GOBIN=/usr/local/bin go install
