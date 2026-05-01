#!/bin/sh
# run_tests.sh
# Executed inside the Go app container (target: test).
# Responsibility: run all Go unit tests.
# Acceptance tests are handled separately by the python-tests container.
set -e

export DB_SHOULD_MIGRATE=true

run_unit_tests() {
  echo "==> Running Go unit tests..."
  go test ./... \
    -coverprofile=coverage/coverage.out \
    -covermode=atomic \
    -count=1
}

run_unit_tests
