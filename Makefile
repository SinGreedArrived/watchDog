.PHONY: build
build:
#		go build -v -gcflags=all="-N -l" ./cmd/parser
	go build -v ./cmd/parser

run: build
	./parser

.PHONY: test
test:
	go test -v -timeout 20s ./...

.DEFAULT_GOAL := build
