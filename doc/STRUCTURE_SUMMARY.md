# Project Structure Summary

## Overview

`torch2pprof` is a unified command-line tool that converts PyTorch profiler traces to pprof format and provides trace analysis capabilities.

## Directory Layout

```
torch2pprof/
├── cmd/                              # Command-line applications
│   └── torch2pprof/
│       └── main.go                   # Unified tool with subcommands
│
├── internal/                         # Private implementation packages
│   ├── profile/
│   │   └── profile.go                # pprof protobuf encoding
│   └── converter/
│       ├── trace.go                  # Trace loading and conversion
│       └── analyzer.go               # Trace analysis
│
├── test/                             # Test data and utilities
│   └── pprof_verification.py         # pprof validation script
│
├── data/                             # Sample data
│   └── profff.json                   # Example PyTorch trace
│
├── Makefile                          # Build automation
├── README.md                         # User documentation
├── SETUP.md                          # Setup and migration guide
├── PROJECT_LAYOUT.md                 # Architecture details
├── go.mod                            # Module definition
├── go.sum                            # Dependency checksums
└── .gitignore                        # Git exclusions
```

## Package Organization

### `internal/profile`
- **Purpose**: pprof protobuf format encoding
- **Main Types**:
  - `Profile` - Complete profile structure
  - `Builder` - Thread-safe construction
  - `Sample`, `Location`, `Function`, `ValueType`, `Line` - Profile components
- **Key Methods**:
  - `(Profile).Encode()` - Convert to protobuf bytes
  - `(Builder).AddString()` - Intern strings
  - `(Builder).SetSampleTypes()` - Configure samples
  - `(Builder).GetOrCreateLocation()` - Create locations

### `internal/converter`
- **Purpose**: PyTorch trace conversion and analysis
- **Main Types**:
  - `TraceEvent` - Single event from trace
  - `TraceData` - Parsed trace file
  - `TraceAnalysis` - Statistics and results
- **Key Functions**:
  - `LoadTraceFile()` - Parse JSON trace
  - `ConvertTrace()` - Convert to pprof profile
  - `AnalyzeTrace()` - Generate statistics
  - `ProcessThreadEvents()` - Per-thread stack building

### `cmd/torch2pprof`
- **Purpose**: Unified CLI tool
- **Subcommands**:
  - `convert` - Convert PyTorch trace to pprof format
  - `analyze` - Analyze trace and show statistics
  - (default) - Convert (backward compatible)
- **CLI Examples**:
  ```bash
  torch2pprof convert trace.json profile.pb.gz
  torch2pprof analyze -top 50 trace.json
  torch2pprof trace.json profile.pb.gz  # backward compatible
  ```

## Module Path

**Module Name**: `pytorch-to-pprof` (from `go.mod`)

All imports use the full path:
```go
import "pytorch-to-pprof/internal/profile"
import "pytorch-to-pprof/internal/converter"
```

## Building

### Basic Commands
```bash
make build          # Compile torch2pprof to bin/
make install        # Install to $GOPATH/bin
make clean          # Remove build artifacts
make test           # Run tests
make fmt            # Format code
make vet            # Check code quality
make dist           # Build for multiple platforms
```

### Output Locations
- **Development builds**: `bin/torch2pprof`
- **Distribution builds**: `dist/torch2pprof-*`

## Dependencies

Minimal external dependencies:
- `github.com/google/pprof` (indirect, for reference)
- `google.golang.org/protobuf` (indirect, for reference)

No explicit dependencies in `go.mod` - all code is self-contained.

## Architecture Changes

### Before
```
Two separate binaries:
├── torch2pprof               # Converter only
└── analyze-pytorch           # Analyzer only
```

### After
```
One unified tool:
└── torch2pprof
    ├── convert               # Subcommand
    ├── analyze               # Subcommand
    └── (default)             # Backward compatible convert
```

## Key Improvements

| Aspect | Old | New |
|--------|-----|-----|
| **Binaries** | 2 separate | 1 unified tool |
| **Interface** | Different CLIs | Consistent subcommands |
| **Installation** | 2 binaries | 1 binary |
| **Distribution** | 2 files/platform | 1 file/platform |
| **Maintenance** | 2 entry points | 1 entry point |
| **User Experience** | Learn 2 tools | Learn 1 tool |
| **Backward Compat** | N/A | Old syntax works |

## Verification

Tool builds and runs successfully:

```bash
$ make build
Build complete: bin/torch2pprof

$ ./bin/torch2pprof convert data/profff.json /tmp/test.pb.gz
Success!
  - 68571 samples
  - 1003 locations
  - 1003 functions

$ ./bin/torch2pprof analyze data/profff.json
PyTorch Profile Analysis
Total events:           1254248
Unique operations:      920
Total time:             109.083 s

# Backward compatible
$ ./bin/torch2pprof data/profff.json /tmp/test.pb.gz
Success!
```

## Usage Examples

### Convert a PyTorch trace
```bash
# Generate in PyTorch
python -m torch.profiler export_trace_handler my_trace.json

# Convert to pprof
torch2pprof convert my_trace.json profile.pb.gz

# Analyze with pprof
go tool pprof profile.pb.gz
```

### Analyze a trace before converting
```bash
# Get overview
torch2pprof analyze my_trace.json

# See top 100 operations
torch2pprof analyze -top 100 my_trace.json
```

### Backward compatible workflow
```bash
# Old syntax still works
torch2pprof my_trace.json profile.pb.gz
```

## Commands Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `convert` | Convert trace → pprof | `torch2pprof convert in.json out.pb.gz` |
| `analyze` | Show statistics | `torch2pprof analyze -top 50 in.json` |
| (default) | Convert (compat) | `torch2pprof in.json out.pb.gz` |

## Next Steps

1. **Add Tests**: Create `*_test.go` files in `internal/` packages
2. **Add Examples**: Create `examples/` directory with sample usage
3. **CI/CD**: Set up GitHub Actions for automated building/testing
4. **Documentation**: Add godoc comments to exported functions
5. **New Subcommands**: Consider adding `diff`, `merge`, `filter` subcommands

## Documentation Files

- **README.md** - User guide and overview
- **PROJECT_LAYOUT.md** - Detailed architecture explanation
- **SETUP.md** - Migration guide and setup instructions
- **This file** - Quick reference summary

## Maintenance

- Code follows Go conventions and is formatted with `go fmt`
- Quality checked with `go vet`
- Structure allows for future expansion
- Clear separation between public tools and private implementation
- Unified CLI provides consistent user experience
