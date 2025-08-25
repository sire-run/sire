.PHONY: test lint lint-fix coverage

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