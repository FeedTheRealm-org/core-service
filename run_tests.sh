PATTERN_FOR_ACCEPTANCE_TESTS="./internal/*/acceptance-tests/*_test.go"

export DB_SHOULD_MIGRATE=false

run_unit_tests() {
  go test ./... -cover
}

run_acceptance_tests() {
  for test_file in $PATTERN_FOR_ACCEPTANCE_TESTS; do
    echo "Testing $test_file"
    if [ -f "$test_file" ]; then
      (cd "$(dirname "$test_file")" && go test -v --godog.tags=~wip)
    fi
  done
}

run_unit_tests
run_acceptance_tests
