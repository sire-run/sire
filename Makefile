.PHONY: test lint lint-fix coverage format

test:
	@echo "Running all Go tests..."
	go test ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run

lint-fix:
	@echo "Running golangci-lint with --fix..."
	golangci-lint run --fix

coverage:
	@echo "Generating test coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	@echo "Coverage report generated: coverage.out (HTML: coverage.html)"

format:
	@echo "🎨 Applying code formatters..."
	@echo "  - Standard Go formatting..."
	@gofmt -w .
	@echo "  - Organizing imports..."
	@goimports -w .
	@echo "  - Strict formatting with gofumpt..."
	@gofumpt -w . 2>/dev/null || echo "    (gofumpt not available, skipping)"
	@echo "✅ Code formatting complete"
