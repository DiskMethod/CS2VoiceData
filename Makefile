.PHONY: fmt fmt-check vet check build

# Format code using goimports
fmt:
	goimports -w .

# Check formatting without writing changes (fail if unformatted)
fmt-check:
	@UNFORMATTED=$$(goimports -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi

# Run go vet on all packages
vet:
	go vet ./...

# Run all checks (formatting and vet)
check: fmt-check vet

# Build all packages with CGO enabled by default
CGO_ENABLED ?= 1
build:
	CGO_ENABLED=$(CGO_ENABLED) go build ./...
