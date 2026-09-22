.PHONY: run tidy build test

-include .env

CONTROLLER_SECRET ?= dev-secret
TLS_CERT ?=
TLS_KEY ?=
ifeq ($(TLS_CERT),)
ADDR ?= :8080
else
ADDR ?= :443
endif

tidy:
	go mod tidy

build:
	go build -o bin/eos-page ./cmd/server

test:
	go test ./...

run:
	CONTROLLER_SECRET=$(CONTROLLER_SECRET) ADDR=$(ADDR) TLS_CERT=$(TLS_CERT) TLS_KEY=$(TLS_KEY) go run ./cmd/server
