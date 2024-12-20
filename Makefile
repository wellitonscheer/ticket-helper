include .env
export

APP_BINARY_NAME=ticket-helper

.PHONY: help
help: ## display this help message
	@echo "Usage: make <target>"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-26s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## execute everything it needs to run dev
	go install github.com/air-verse/air@latest
	chmod +x standalone_embed.sh embedding.sh attu.sh

.PHONY: dev
dev: ## run everything it needs to start in dev with hot reload
	./standalone_embed.sh start
	./embedding.sh
	./attu.sh
	air
