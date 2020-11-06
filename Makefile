.PHONY: build
build:
#		go build -v -gcflags=all="-N -l" ./cmd/parser
	go build -v ./cmd/parser

windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -v -o parser_cgo1.exe ./cmd/parser 

run: build
	./parser

.PHONY: test
test:
	go test -v -timeout 20s ./...

.DEFAULT_GOAL := build
