.PHONY: build
build:
#		go build -v -gcflags=all="-N -l" ./cmd/parser
	go build -v ./cmd/parser

release: build
	sudo cp parser /usr/local/bin/parser
	cp configs/config.toml /home/greed/.config/parser/config.toml

run: build
	./parser

.PHONY: test
test:
	go test -v -timeout 20s ./...

.DEFAULT_GOAL := build
