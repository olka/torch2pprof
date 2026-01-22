# Testing Guide

## Overview

torch2pprof has comprehensive unit tests with high code coverage and CI/CD integration.

## Test Statistics

- **Total Tests**: 20
- **Coverage**: 
  - `internal/converter`: 96.2%
  - `internal/profile`: 93.0%
- **Race Detection**: Enabled
- **Platforms Tested**: Linux, macOS, Windows

## Running Tests

### Basic Tests

```bash
# Run all tests
make test

# Or directly with go
go test ./...
```

### Tests with Coverage

```bash
# Run tests with coverage report
make test-coverage

# This generates:
# - coverage.out (coverage data)
# - coverage.html (HTML report)
```

Open `coverage.html` in your browser to see detailed coverage.

### Tests with Race Detector

```bash
# Run tests with race detector
make test-race

# Or directly
go test -race ./...
```

### Verbose Tests

```bash
# See detailed test output
go test -v ./...

# Test specific package
go test -v ./internal/profile/
go test -v ./internal/converter/
```

## Test Structure

### Profile Package Tests (`internal/profile/profile_test.go`)

Tests for pprof profile encoding:

- `TestNewBuilder` - Builder initialization
- `TestAddString` - String interning
- `TestGetOrCreateFunction` - Function deduplication
- `TestGetOrCreateLocation` - Location creation
- `TestSetSampleTypes` - Sample type configuration
- `TestSetPeriodType` - Period type configuration
- `TestProfileEncode` - Protobuf encoding
- `TestEncodeVarint` - Varint encoding
- `TestConcurrentAccess` - Thread safety
- `TestBuild` - Profile building

### Converter Package Tests (`internal/converter/converter_test.go`)

Tests for trace conversion and analysis:

- `TestGetTid` - Thread ID extraction
- `TestLoadTraceFile_PlainJSON` - Loading plain JSON
- `TestLoadTraceFile_GzipJSON` - Loading compressed JSON
- `TestLoadTraceFile_GzipWithoutExtension` - Magic number detection
- `TestLoadTraceFile_NonexistentFile` - Error handling
- `TestAnalyzeTrace` - Trace analysis
- `TestGetSortedCategories` - Category sorting
- `TestGetSortedOperations` - Operation sorting
- `TestConvertTrace` - Trace to pprof conversion
- `TestConvertTrace_EmptyEvents` - Empty trace handling
- `TestConvertTrace_FilteredEvents` - Event filtering

## Test Coverage

### Viewing Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage by package
go test -cover ./...

# View detailed coverage
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Goals

- Target: >90% coverage for all packages
- Current: 96.2% (converter), 93.0% (profile)

### What's Tested

✅ **Tested**:
- Profile building and encoding
- String interning and deduplication
- Function and location management
- Thread safety (concurrent access)
- Trace file loading (plain and gzip)
- Automatic compression detection
- Trace analysis and statistics
- Event filtering
- Error handling

❌ **Not Tested** (by design):
- CLI argument parsing (cmd/torch2pprof)
- Progress reporting
- User interaction

## Continuous Integration

### GitHub Actions

The project uses GitHub Actions for CI/CD with three workflows:

#### 1. CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request:

**Test Job**:
- Tests on Linux, macOS, Windows
- Tests with Go 1.23 and 1.24
- Runs `go vet`, `go fmt`, tests with race detector
- Generates coverage reports
- Uploads coverage to Codecov

**Build Job**:
- Builds binary
- Tests basic functionality
- Uploads artifacts

**Lint Job**:
- Runs golangci-lint
- Checks code quality

#### 2. Release Workflow (`.github/workflows/release.yml`)

Triggered on version tags:

- Builds binaries for all platforms
- Creates GitHub release
- Uploads binaries with checksums
- Tests binaries on target platforms

### CI Status

Add these badges to your README:

```markdown
[![CI](https://github.com/yourusername/torch2pprof/workflows/CI/badge.svg)](https://github.com/yourusername/torch2pprof/actions)
[![codecov](https://codecov.io/gh/yourusername/torch2pprof/branch/main/graph/badge.svg)](https://codecov.io/gh/yourusername/torch2pprof)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/torch2pprof)](https://goreportcard.com/report/github.com/yourusername/torch2pprof)
```

## Writing New Tests

### Test File Naming

- Place tests in the same package: `*_test.go`
- Profile tests: `internal/profile/profile_test.go`
- Converter tests: `internal/converter/converter_test.go`

### Test Function Naming

```go
func TestFunctionName(t *testing.T) {
    // Test implementation
}

func TestFunctionName_SpecificCase(t *testing.T) {
    // Test specific case
}
```

### Table-Driven Tests

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"case1", "input1", 1},
        {"case2", "input2", 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("expected %d, got %d", tt.expected, result)
            }
        })
    }
}
```

### Using t.TempDir()

For tests requiring temporary files:

```go
func TestWithTempFile(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    testFile := filepath.Join(tmpDir, "test.json")
    
    // Write test data
    os.WriteFile(testFile, data, 0644)
    
    // Test
    result, err := LoadFile(testFile)
    // ...
}
```

## Benchmarking

### Running Benchmarks

```bash
# Run benchmarks
go test -bench=. ./...

# With memory stats
go test -bench=. -benchmem ./...

# Specific package
go test -bench=. ./internal/profile/
```

### Writing Benchmarks

```go
func BenchmarkEncode(b *testing.B) {
    profile := setupProfile()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        profile.Encode()
    }
}
```

## Integration Testing

### Manual Integration Tests

```bash
# Test with real data
./bin/torch2pprof convert data/trace.json.gz /tmp/test.pb.gz
./bin/torch2pprof analyze data/trace.json.gz

# Verify output
go tool pprof /tmp/test.pb.gz
```

### Automated Integration Tests

Consider adding to CI:

```bash
#!/bin/bash
# integration_test.sh

set -e

# Build
make build

# Test convert
./bin/torch2pprof convert testdata/sample.json /tmp/out.pb.gz

# Verify output exists and is gzipped
file /tmp/out.pb.gz | grep gzip

# Test analyze
./bin/torch2pprof analyze testdata/sample.json | grep "Total events"

echo "Integration tests passed!"
```

## Debugging Tests

### Verbose Output

```bash
# See all test output
go test -v ./...

# Run specific test
go test -v -run TestSpecificTest ./internal/profile/
```

### Test with dlv (Delve Debugger)

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/profile/ -- -test.run TestSpecificTest
```

### Print Debugging

```go
func TestDebug(t *testing.T) {
    result := Function()
    t.Logf("Result: %+v", result) // Only shown with -v
    // ...
}
```

## Best Practices

1. **Test One Thing**: Each test should verify one behavior
2. **Use Subtests**: Group related tests with `t.Run()`
3. **Clean Up**: Use `t.Cleanup()` or `defer` for cleanup
4. **Table-Driven**: Use tables for multiple similar cases
5. **Descriptive Names**: Test names should describe what they test
6. **Don't Skip**: Avoid `t.Skip()` without good reason
7. **Fast Tests**: Keep tests fast (<1s per test)
8. **Deterministic**: Tests should produce same results every time
9. **Coverage**: Aim for >90% but don't chase 100%
10. **Race Detection**: Always run with `-race` before committing

## Troubleshooting

### Tests Fail Intermittently

Likely a race condition:
```bash
go test -race -count=100 ./...
```

### Tests Fail on CI but Pass Locally

Check:
- OS-specific file paths (use `filepath.Join`)
- Line endings (Windows vs Unix)
- Environment variables
- Available resources

### Coverage Too Low

```bash
# Find untested code
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Test Examples](https://go.dev/doc/tutorial/add-a-test)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Advanced Testing](https://about.sourcegraph.com/blog/go/advanced-testing-in-go)
