SHELL := /usr/bin/env bash -e

all: deps test

clean:
	go clean

init:
	docker-compose up -d

nuke: clean
	docker-compose down

deps: clean
	glide install

test: nuke init
	go test -v -cover ./...
