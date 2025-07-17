#!/bin/bash
set -e

echo "Running all Go tests in the project..."

# Run tests and capture output
test_output=$(go test ./... -v | tee /dev/tty)

# Count total tests and passes
total=$(echo "$test_output" | grep -E "^=== RUN" | wc -l)
passes=$(echo "$test_output" | grep -E "^--- PASS" | wc -l)

echo "Passed: $passes / $total tests"
echo "All tests completed."
