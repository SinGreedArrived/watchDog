build: test
#		go build -v -gcflags=all="-N -l" ./cmd/parser
	go build -v ./cmd/parser

release: build
	sudo cp parser /usr/local/bin/parser
#	cp configs/config.toml /home/greed/.config/parser/config.toml

run: build
	./parser

test:
	go test -v -timeout 10s ./...

.DEFAULT_GOAL := build
.PHONY: test build release run
