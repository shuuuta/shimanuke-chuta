.PHONY: help
.DEFAULT_GOAL := help

.PHONY: serve
serve: ## Run dev server
	go run ./tool serve --http localhost:8080 --no-open

.PHONY: servefull
servefull: ## Run dev server open to local network
	go run ./tool serve --http 0.0.0.0:8080 --no-open

.PHONY: build
build: ## Build for dev server
	go run ./tool build

.PHONY: builddev
builddev: ## Build for dev server with dev status
	go run ./tool build -tags=dev

## ------------------
help: ## Show options
	@grep -E '^[a-zA-Z][a-zA-Z0-9]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
