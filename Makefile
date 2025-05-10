.PHONY: fmt fmt-check vet check

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
