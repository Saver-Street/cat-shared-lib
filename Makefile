.PHONY: test test-v test-race lint cover cover-html check-coverage bench fuzz clean help

.DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests
	go test ./... -count=1

test-v: ## Run unit tests (verbose)
	go test ./... -v -count=1

test-race: ## Run tests with race detector
	go test ./... -race -count=1

lint: ## Run linters (golangci-lint)
	golangci-lint run --timeout 120s ./...

cover: ## Generate coverage report
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

cover-html: cover ## Open coverage report in browser
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

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
	rm -f coverage.out coverage.html

bench: ## Run all benchmarks
	go test ./... -bench=. -benchmem -count=1 -run=^$$ -timeout 120s

fuzz: ## Run all fuzz tests (short smoke run per target)
	@echo "Running fuzz tests (5s per target)..."
	@failed=0; \
	for pkg in $$(go list ./... 2>/dev/null); do \
		for target in $$(go test $$pkg -list '^Fuzz' 2>/dev/null | grep '^Fuzz' || true); do \
			echo "  $$pkg $$target"; \
			go test $$pkg -fuzz="^$${target}$$" -fuzztime=5s -run='^$$' 2>&1 | tail -1; \
			if [ $$? -ne 0 ]; then failed=$$((failed+1)); fi; \
		done; \
	done; \
	if [ $$failed -ne 0 ]; then echo "FAIL: $$failed fuzz target(s) failed"; exit 1; fi; \
	echo "OK: all fuzz targets passed"
