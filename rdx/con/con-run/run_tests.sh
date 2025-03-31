#!/bin/bash

# Download dependencies
go mod tidy

# Run all tests with verbose output and code coverage
go test -v ./pkg/... -cover

# If you want to generate a coverage report, uncomment these lines:
# go test ./pkg/... -coverprofile=coverage.out
# go tool cover -html=coverage.out -o coverage.html
# echo "Coverage report generated at coverage.html"