.PHONY: test test-v test-race lint cover check-coverage clean help

.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests
	go test ./... -count=1

test-v: ## Run unit tests (verbose)
	go test ./... -v -count=1

test-race: ## Run tests with race detector
	go test ./... -race -count=1

lint: ## Run linters (vet + staticcheck)
	go vet ./...
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

cover: ## Generate coverage report
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

check-coverage: cover ## Fail if total coverage < 95%
	@coverage=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	threshold=95; \
	if [ "$${coverage%.*}" -lt "$$threshold" ]; then \
		echo "FAIL: coverage $${coverage}% is below $${threshold}%"; \
		exit 1; \
	else \
		echo "OK: coverage $${coverage}% meets $${threshold}% threshold"; \
	fi

clean: ## Remove build artifacts
	rm -f coverage.out
