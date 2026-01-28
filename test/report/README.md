# Test Reports

This directory contains test execution reports and logs.

## Generated Files

- `test_e2e_out.log` - E2E test execution output
- `test_out.log` - General test execution output
- `test_output.txt` - Test output in text format
- `test_results.log` - Detailed test results

## Note

These files are auto-generated during test execution and are excluded from version control via `.gitignore`.

## Usage

To generate test reports, run:

```bash
# Run integration tests
go test -v ./test/integration/... > test/report/test_output.txt 2>&1

# Run specific package tests
go test -v ./test/integration/minio/... > test/report/minio_test.log 2>&1
go test -v ./test/integration/user/client/... > test/report/client_test.log 2>&1
go test -v ./test/integration/user/session/... > test/report/session_test.log 2>&1
```
