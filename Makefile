SHELL := /usr/bin/env bash -e

all: deps test

clean:
	go clean

deps: clean
	glide install

test:
	go test ./...
