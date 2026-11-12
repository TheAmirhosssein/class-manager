.PHONY: build run serve dev prod seed migrate migrate-down test tidy air

BINARY := bin/classino
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

build:
	go build -o $(BINARY) ./cmd

run: build
	./$(BINARY) serve

serve:
	go run ./cmd serve

# Backend (air if installed) + frontend preview with /api proxy — see scripts/dev.sh
dev:
	@bash scripts/dev.sh

# Production: build binary, load root .env, serve API + static UI on HTTP_PORT
prod:
	@bash scripts/prod.sh

seed:
	go run ./cmd seed

migrate:
	go run ./cmd migrate

migrate-down:
	go run ./cmd migrate --down

test:
	go test ./...

tidy:
	go mod tidy

air:
	air
