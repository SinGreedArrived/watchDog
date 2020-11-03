.PHONY: build
build:
#		go build -v -gcflags=all="-N -l" ./cmd/parser
	go build -v ./cmd/parser

.DEFAULT_GOAL := build

