#!/bin/sh
# run_tests.sh
# Executed inside the Go app container (target: test).
# Responsibility: run all Go unit tests.
# Acceptance tests are handled separately by the python-tests container.

export DB_SHOULD_MIGRATE=false

run_unit_tests() {
  echo "==> Running Go unit tests..."
  go test ./... -cover
}

run_unit_tests
