.PHONY: build
build:
	go build -v -gcflags=all="-N -l" ./cmd/parser

.DEFAULT_GOAL := build

