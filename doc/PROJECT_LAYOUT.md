# Project Layout

This document explains the structure of the torch2pprof project following Go conventions.

## Directory Structure

```
torch2pprof/
├── cmd/                          # Command-line applications
│   └── torch2pprof/              # Main tool with subcommands
│       └── main.go               # Entry point with convert & analyze commands
│
├── internal/                     # Private packages (not for external import)
│   ├── profile/
│   │   └── profile.go            # pprof protobuf encoding
│   └── converter/                # Core conversion and analysis logic
│       ├── trace.go              # Trace loading, processing, and conversion
│       └── analyzer.go           # Trace analysis and statistics
│
├── test/                         # Test data and utilities
│   └── pprof_verification.py     # Python script to verify pprof output
│
├── data/                         # Sample data
│   └── trace.json.gz             # Example PyTorch trace
│
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── Makefile                      # Build automation
├── README.md                     # User documentation
```

## Principles

This project follows standard Go project layout conventions:

### `/cmd` - Command-line Tools

The `cmd/torch2pprof/` directory contains a single executable with multiple subcommands:
- `convert` - Convert PyTorch trace to pprof format
- `analyze` - Analyze trace and show statistics

Main file delegates to internal packages and provides CLI interface.

### `/internal` - Private Packages

Code in the `internal/` directory cannot be imported by external packages. This enforces clear boundaries:

#### `internal/profile/`
- **Responsibility**: pprof profile structure and encoding
- **Exports**: 
  - `Profile` - Main profile structure
  - `Builder` - Thread-safe profile builder
  - `ValueType`, `Sample`, `Location`, `Function`, `Line` - Profile components
- **Key functions**:
  - `(Profile).Encode()` - Protobuf encoding
  - `(Builder).Build()` - Finalize profile construction

#### `internal/converter/`
- **Responsibility**: Trace conversion and analysis
- **Package**: `converter`
- **Exports**:
  - `TraceData` - Parsed trace structure
  - `TraceEvent` - Individual event
  - `TraceAnalysis` - Analysis results
  - `LoadTraceFile()` - Load JSON trace
  - `ConvertTrace()` - Convert to pprof
  - `AnalyzeTrace()` - Analyze statistics
- **Key internal functions**:
  - `ProcessThreadEvents()` - Stack-based event processing
  - `getTid()` - Thread ID extraction
  - `ConvertOptions` - Conversion configuration

### `/cmd/torch2pprof` - Unified Tool

The tool uses a subcommand architecture:
- **Main function**: Routes to appropriate subcommand
- **convertCommand()**: Handles trace conversion
- **analyzeCommand()**: Handles trace analysis
- **Backward compatibility**: Default behavior converts when no subcommand specified

## Module Organization

**Module name**: `pytorch-to-pprof`

All imports use the full module path:
```go
import "pytorch-to-pprof/internal/profile"
import "pytorch-to-pprof/internal/converter"
```

## Import Rules

1. **From `cmd/torch2pprof`**: May import from `internal/`
2. **From `internal/profile`**: May import only standard library
3. **From `internal/converter`**: May import `internal/profile` and standard library
4. **External packages**: Only imported via `internal/` packages

This creates a clean dependency hierarchy:
```
cmd/torch2pprof
    ├── internal/converter
    │   └── internal/profile
    └── internal/profile
```

## Subcommand Architecture

The tool uses Go's standard approach for subcommands:

```go
func main() {
    switch os.Args[1] {
    case "convert":
        convertCommand(os.Args[2:])
    case "analyze":
        analyzeCommand(os.Args[2:])
    default:
        // Backward compatibility: default to convert
        convertCommand(os.Args[1:])
    }
}
```

Each subcommand:
- Has its own `flag.FlagSet` for independent argument parsing
- Provides specific usage documentation
- Delegates to internal packages for implementation

## Adding New Code

### New Subcommand

1. Add case to main switch statement
2. Create `<name>Command()` function
3. Set up `flag.FlagSet` for arguments
4. Import from `internal/` as needed
5. Update help text and README

### New Core Functionality

1. Create file in appropriate `internal/` package
2. Use unexported functions for helpers
3. Document exported types and functions
4. Add tests alongside implementation

### New Package

Only create new packages if functionality doesn't fit existing ones:
1. Place in `internal/` to keep it private
2. Choose a descriptive name (e.g., `internal/metrics`, `internal/output`)
3. Document the package's responsibility
4. Update this layout document

## Testing

Tests follow Go conventions:
- `*_test.go` files in the same package
- Black-box testing (import package normally, not via `internal`)
- Test files can access unexported functions in the same package

Currently, test data is in `test/` with external test utilities.

## Building

See `Makefile` for build targets:
- `make build` - Compile torch2pprof binary
- `make install` - Install to `$GOPATH/bin`
- `make clean` - Remove build artifacts
- `make dist` - Build for multiple platforms

Binary is placed in `bin/` directory.

## Dependencies

See `go.mod` and `go.sum` for the full dependency list. Currently minimal:
- `github.com/google/pprof` (indirect)
- `google.golang.org/protobuf` (indirect)

All imports are in `internal/` packages to keep dependencies centralized.

## Commands Overview

| Command | Purpose | Example |
|---------|---------|---------|
| `convert` | Convert trace to pprof | `torch2pprof convert trace.json out.pb.gz` |
| `analyze` | Show trace statistics | `torch2pprof analyze -top 50 trace.json` |
| (default) | Convert (compatibility) | `torch2pprof trace.json out.pb.gz` |

## Future Improvements

Possible enhancements to the structure:
- `pkg/` directory if the library should be importable
- `examples/` for usage examples
- `docs/` for detailed documentation
- `scripts/` for utility scripts
- `testdata/` for test fixtures
- Additional subcommands (e.g., `diff`, `merge`, `filter`)
