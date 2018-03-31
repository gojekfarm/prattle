SHELL := /usr/bin/env bash -e

all: deps test

clean:
	go clean

init:
	docker-compose up -d

nuke: clean
	docker-compose down

deps: clean
	dep ensure

test: nuke init
	go test -v -cover ./...
