SHELL := /usr/bin/env bash -e

all: deps test

clean:
	go clean

deps:
	glide install

test:
	go test
