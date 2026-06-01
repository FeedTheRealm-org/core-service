# Testing Guide

This section describes how we test the core-service. Tests should be co-located with the tested code, appending `_test` to the package name.

## Running Tests

To run all tests including acceptance tests inside an isolated docker environment:

```bash
make test
```

To run only Go unit tests:

```bash
make test-unit
```

To run only Python/behave acceptance tests:

```bash
make test-acceptance
```
