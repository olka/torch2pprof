# Setup Guide

This document explains the project structure and how to work with it.

## Project Structure

The project has been organized following Go best practices:

```
cmd/
└── torch2pprof/
    └── main.go              # Unified tool with subcommands

internal/
├── profile/
│   └── profile.go           # pprof profile encoding
└── converter/
    ├── trace.go             # Trace conversion logic
    └── analyzer.go          # Trace analysis logic
```

## Tool Commands

torch2pprof is a unified tool with two subcommands:

### Convert Command

Convert PyTorch traces to pprof format:

```bash
# Using the convert subcommand
torch2pprof convert input.json output.pb.gz

# Or use default behavior (for backward compatibility)
torch2pprof input.json output.pb.gz
```

### Analyze Command

Analyze traces and show statistics:

```bash
# Show top 20 operations (default)
torch2pprof analyze trace.json

# Show top 50 operations
torch2pprof analyze -top 50 trace.json
```

## What Changed

### From Multiple Binaries to Single Tool

**Before**: Two separate binaries
- `torch2pprof` - Converter
- `analyze-pytorch` - Analyzer

**After**: One unified tool with subcommands
- `torch2pprof convert` - Converter
- `torch2pprof analyze` - Analyzer
- `torch2pprof` (default) - Converter (backward compatible)

### Benefits

- **Unified Interface**: One tool to install and manage
- **Consistent UX**: Standard subcommand pattern (like `git`, `docker`)
- **Easier Distribution**: Single binary to distribute
- **Simpler Installation**: One tool in PATH
- **Backward Compatible**: Old `torch2pprof <in> <out>` still works

## Code Organization

### Profile Encoding (`internal/profile/profile.go`)
- `Profile`, `Builder`, `ValueType`, `Sample`, `Location`, `Function`, `Line` types
- Protobuf encoding methods
- Thread-safe `Builder` with string interning

### Trace Conversion (`internal/converter/trace.go`)
- `TraceEvent`, `TraceData` types
- `LoadTraceFile()` - Load JSON traces
- `ConvertTrace()` - Main conversion algorithm
- `ProcessThreadEvents()` - Per-thread processing

### Trace Analysis (`internal/converter/analyzer.go`)
- `TraceAnalysis` - Analysis results
- `AnalyzeTrace()` - Generate statistics
- Helper types for sorting results

### Main Tool (`cmd/torch2pprof/main.go`)
- Subcommand routing
- `convertCommand()` - Conversion logic
- `analyzeCommand()` - Analysis logic
- Backward compatibility handling

## Building

### Quick Start

```bash
# Build the tool
make build

# Tool is in bin/ directory
./bin/torch2pprof convert data/profff.json output.pb.gz
./bin/torch2pprof analyze data/profff.json
```

### Development

```bash
# Format code
make fmt

# Check for issues
make vet

# Build and check
make dev

# Install to $GOPATH/bin
make install
```

### Distribution

```bash
# Build for Linux, macOS, Windows
make dist

# Binaries are in dist/ directory
```

## Migration from Old Structure

If you were using the old two-binary setup:

### Old Commands
```bash
torch2pprof trace.json output.pb.gz
analyze-pytorch trace.json
```

### New Commands (Recommended)
```bash
torch2pprof convert trace.json output.pb.gz
torch2pprof analyze trace.json
```

### Backward Compatible (Still Works)
```bash
torch2pprof trace.json output.pb.gz  # Automatically uses convert
```

## Module Path

The module is named `pytorch-to-pprof` as defined in `go.mod`. All imports use:

```go
import "pytorch-to-pprof/internal/profile"
import "pytorch-to-pprof/internal/converter"
```

## Testing

Create test files alongside implementation:

```bash
internal/profile/profile_test.go    # Tests for profile package
internal/converter/trace_test.go    # Tests for converter package
```

Run tests with:
```bash
make test
```

## Adding Features

### New Subcommand

Add a new subcommand to the tool:

1. Add case in `main()` switch statement:
```go
case "newcommand":
    newCommand(os.Args[2:])
```

2. Implement the command function:
```go
func newCommand(args []string) {
    fs := flag.NewFlagSet("newcommand", flag.ExitOnError)
    // ... setup flags
    fs.Parse(args)
    // ... implementation
}
```

3. Update help text in `printUsage()`
4. Update README.md

### New Functionality

1. Add to appropriate `internal/` package
2. Export types/functions if needed by other packages
3. Keep implementation details unexported
4. Add tests

## Project Files

- `README.md` - User documentation
- `PROJECT_LAYOUT.md` - Structure explanation
- `SETUP.md` - This file
- `Makefile` - Build automation
- `go.mod` / `go.sum` - Dependencies
- `.gitignore` - Version control exclusions

## Key Improvements

| Aspect | Before | After |
|--------|--------|-------|
| Binaries | 2 separate tools | 1 unified tool |
| Commands | Different programs | Subcommands |
| Installation | Install 2 binaries | Install 1 binary |
| Distribution | 2 files per platform | 1 file per platform |
| User Experience | Inconsistent | Unified interface |
| Backward Compat | N/A | Old syntax still works |

## Common Workflows

### Convert and Analyze

```bash
# Convert trace
torch2pprof convert trace.json profile.pb.gz

# Analyze original trace
torch2pprof analyze trace.json

# View converted profile
go tool pprof profile.pb.gz
```

### Quick Analysis

```bash
# See what's in the trace before converting
torch2pprof analyze -top 100 trace.json
```

### Backward Compatible

```bash
# Old way (still works)
torch2pprof trace.json profile.pb.gz
```

## Next Steps

1. Add unit tests in `internal/*/` directories
2. Add integration tests for subcommands
3. Consider adding more subcommands (e.g., `diff`, `merge`, `filter`)
4. Add GitHub Actions CI/CD if using GitHub
5. Document API with godoc comments

## Questions?

Refer to:
- `README.md` - Usage guide
- `PROJECT_LAYOUT.md` - Architecture
- `Makefile` - Build targets
- Code comments in `internal/` packages
