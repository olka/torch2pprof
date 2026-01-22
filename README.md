# torch2pprof

A tool to convert PyTorch profiler traces to the pprof format for visualization with `go tool pprof`.

## Overview

PyTorch's profiler outputs traces in Chrome Trace Event format (JSON), which is difficult to analyze directly. This tool converts those traces into pprof's binary format, allowing you to:

- Visualize call stacks
- Identify performance bottlenecks
- Analyze CPU usage patterns
- Use the full suite of pprof analysis tools

## Installation

### From Source

```bash
git clone https://github.com/yourusername/torch2pprof
cd torch2pprof
make install
```

### Using Go

```bash
go install github.com/yourusername/torch2pprof/cmd/torch2pprof@latest
```

## Usage

### Converting PyTorch Traces to pprof

```bash
# Using the convert subcommand
torch2pprof convert input_trace.json output_profile.pb.gz

# Works with compressed files too
torch2pprof convert input_trace.json.gz output_profile.pb.gz

# Or use the default behavior (for backward compatibility)
torch2pprof input_trace.json output_profile.pb.gz
```

This will:
1. Load the PyTorch trace JSON file (supports both `.json` and `.json.gz` files)
2. Parse all complete events (ph=X) with positive durations
3. Build call stacks by analyzing event nesting
4. Encode to pprof protobuf format with gzip compression

**Note**: Input files can be either plain JSON or gzip-compressed. The tool automatically detects compression based on file extension (`.gz`) or file content (magic number detection).

### Analyzing Traces

```bash
# Show top 20 operations (default)
torch2pprof analyze input_trace.json

# Works with compressed files
torch2pprof analyze input_trace.json.gz

# Show top 50 operations
torch2pprof analyze -top 50 input_trace.json.gz
```

This displays:
- Total number of events and statistics
- Time breakdown by category
- Top operations by total time

**Note**: Both `.json` and `.json.gz` files are supported.

### Viewing with pprof

After conversion, analyze the profile with go tool pprof:

```bash
go tool pprof output_profile.pb.gz
```

Common pprof commands:
- `top` - Show top functions by time
- `list <function>` - Show source code with annotations
- `web` - Generate a graph visualization (requires graphviz)
- `flame` - Generate flame graph

## Commands

### convert

Convert PyTorch trace to pprof format.

```bash
torch2pprof convert <input.json|input.json.gz> <output.pb.gz>
```

**Arguments:**
- `input.json|input.json.gz` - PyTorch trace file in Chrome Trace Event format (plain or gzip-compressed)
- `output.pb.gz` - Output pprof profile (gzip compressed)

**Features:**
- Automatically detects gzip compression via `.gz` extension or magic number
- Supports both plain JSON and compressed JSON files

### analyze

Analyze PyTorch trace and show statistics.

```bash
torch2pprof analyze [options] <input.json|input.json.gz>
```

**Options:**
- `-top N` - Show top N operations (default: 20)

**Arguments:**
- `input.json|input.json.gz` - PyTorch trace file to analyze (plain or gzip-compressed)

**Features:**
- Automatically detects gzip compression via `.gz` extension or magic number
- Supports both plain JSON and compressed JSON files

## Project Structure

```
torch2pprof/
├── cmd/
│   └── torch2pprof/        # Main tool with convert & analyze subcommands
├── internal/
│   ├── profile/            # pprof protobuf encoding
│   └── converter/          # Trace loading, conversion, and analysis
└── test/                   # Test data and utilities
```

## Building

```bash
# Build binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Format code
make fmt

# Lint code
make vet

# Build for multiple platforms
make dist
```

## Testing

The project has comprehensive unit tests with high code coverage:

- **Test Coverage**: 96.2% (converter), 93.0% (profile)
- **Total Tests**: 20 unit tests
- **CI/CD**: Automated testing on Linux, macOS, Windows
- **Race Detection**: All tests run with race detector

See [TESTING.md](TESTING.md) for detailed testing documentation.

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage
# Open coverage.html in browser

# Run with race detector
make test-race
```

## How It Works

### Trace Conversion Algorithm

1. **Load Trace**: Parse the JSON trace file containing Chrome Trace Event format
2. **Filter Events**: Keep only complete events (ph=X) with positive duration
3. **Group by Thread**: Organize events by their thread ID
4. **Build Stacks**: For each event, determine its call stack by analyzing event overlaps:
   - Events that temporally contain other events represent parent functions
   - Uses a linear-time stack-based algorithm instead of O(n²) comparison
5. **Aggregate**: Combine identical stacks and sum their durations
6. **Encode**: Convert to pprof protobuf format and compress with gzip

### Performance

- Linear time complexity for stack building (O(n) per thread)
- Parallel processing across multiple threads
- Efficient memory usage with string interning

## Requirements

- Go 1.24 or later

## License

MIT

## Contributing

Contributions are welcome! Please ensure:
- Code passes `go fmt` and `go vet`
- All tests pass
- New features include tests

## Troubleshooting

### Large trace files

For very large trace files (>100MB):
- Ensure sufficient memory (at least 2GB recommended)
- Consider filtering the trace in PyTorch before exporting
- Use gzip-compressed files (`.json.gz`) to save disk space and reduce I/O time
  - Example: A 322MB JSON file compresses to 23MB with gzip (93% reduction)

### Memory usage

The tool maintains maps for:
- String interning (string → index)
- Function deduplication (name+file → ID)
- Location deduplication (name+file → ID)

For profiles with millions of unique functions, this can use several GB.

## Related Tools

- [pprof](https://github.com/google/pprof) - Profile visualization
- [PyTorch Profiler](https://pytorch.org/docs/stable/profiler.html) - Profile generation
- [Chrome DevTools](https://developer.chrome.com/docs/devtools/) - View traces directly
