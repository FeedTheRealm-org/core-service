PATTERN_FOR_ACCEPTANCE_TESTS="./internal/*/acceptance-tests/*_test.go"

run_migrations() {
 	go run ./cmd/migrate/main.go up
}

run_unit_tests() {
  go test ./... -cover
}

run_acceptance_tests() {
  for test_file in $PATTERN_FOR_ACCEPTANCE_TESTS; do
    if [ -f "$test_file" ]; then
      cd "$(dirname "$test_file")" && go test -v --godog.tags=~wip
    fi
  done
}

run_migrations
run_unit_tests
run_acceptance_tests
