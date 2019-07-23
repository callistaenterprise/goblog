SHELL 	 := /bin/bash

fmt:
	find . -name '*.go' | grep -v vendor | xargs gofmt -w -s

test:
	go test ./...

.PHONY: fmt test
