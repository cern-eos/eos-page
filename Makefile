.PHONY: run tidy build test

# Snapshot keys from the shell before .env can replace them with empty values.
_OPENAI_API_KEY := $(OPENAI_API_KEY)
_OPENAI_MODEL := $(OPENAI_MODEL)
_GEMINI_API_KEY := $(GEMINI_API_KEY)
_GOOGLE_API_KEY := $(GOOGLE_API_KEY)

-include .env

ifneq ($(_OPENAI_API_KEY),)
OPENAI_API_KEY := $(_OPENAI_API_KEY)
endif
ifneq ($(_OPENAI_MODEL),)
OPENAI_MODEL := $(_OPENAI_MODEL)
endif
ifneq ($(_GEMINI_API_KEY),)
GEMINI_API_KEY := $(_GEMINI_API_KEY)
endif
ifneq ($(_GOOGLE_API_KEY),)
GOOGLE_API_KEY := $(_GOOGLE_API_KEY)
endif

CONTROLLER_SECRET ?= dev-secret
CHAT_SEARCH_HEADED ?=
TLS_CERT ?=
TLS_KEY ?=
ifeq ($(TLS_CERT),)
ADDR ?= :8080
else
ADDR ?= :443
endif

export CONTROLLER_SECRET
export OPENAI_API_KEY
export OPENAI_MODEL
export GEMINI_API_KEY
export GOOGLE_API_KEY
export CHAT_SEARCH_HEADED
export ADDR
export TLS_CERT
export TLS_KEY

tidy:
	go mod tidy

build:
	go build -o bin/eos-page ./cmd/server

test:
	go test ./...

run:
	go run ./cmd/server
