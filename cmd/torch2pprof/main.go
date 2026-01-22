package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"pytorch-to-pprof/internal/converter"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "convert":
		convertCommand(os.Args[2:])
	case "analyze":
		analyzeCommand(os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
	default:
		// Default behavior for backwards compatibility: convert
		convertCommand(os.Args[1:])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `torch2pprof - PyTorch profiler trace to pprof converter

Usage:
  torch2pprof convert <input.json> <output.pb.gz>   Convert trace to pprof format
  torch2pprof analyze [options] <input.json>        Analyze trace statistics
  torch2pprof <input.json> <output.pb.gz>           Convert (default, for compatibility)

Commands:
  convert     Convert PyTorch trace to pprof format
  analyze     Analyze PyTorch trace and show statistics

Options for analyze:
  -top N      Show top N operations (default: 20)

Examples:
  # Convert trace to pprof
  torch2pprof convert trace.json profile.pb.gz
  torch2pprof trace.json profile.pb.gz

  # Analyze trace
  torch2pprof analyze trace.json
  torch2pprof analyze -top 50 trace.json

`)
}

func convertCommand(args []string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: torch2pprof convert <input.json> <output.pb.gz>\n")
		fmt.Fprintf(os.Stderr, "\nConvert PyTorch profiler trace to pprof format\n")
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	if fs.NArg() != 2 {
		fs.Usage()
		os.Exit(1)
	}

	inputFile := fs.Arg(0)
	outputFile := fs.Arg(1)
	numWorkers := runtime.NumCPU()

	fmt.Printf("Loading %s...\n", inputFile)
	fmt.Printf("Using %d CPU cores\n", numWorkers)

	traceData, err := converter.LoadTraceFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d trace events\n", len(traceData.TraceEvents))

	fmt.Println("Building call stacks (parallel)...")
	start := time.Now()

	profile := converter.ConvertTrace(traceData, converter.ConvertOptions{
		NumWorkers: numWorkers,
	})

	elapsed := time.Since(start)
	fmt.Printf("Conversion complete in %.2fs\n", elapsed.Seconds())

	fmt.Println("Encoding profile...")
	profileBytes, err := profile.Encode()
	if err != nil {
		fmt.Printf("Error encoding profile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Writing to %s...\n", outputFile)
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}

	gz := gzip.NewWriter(f)
	if _, writeErr := gz.Write(profileBytes); writeErr != nil {
		_ = f.Close()
		fmt.Printf("Error writing profile: %v\n", writeErr)
		os.Exit(1)
	}
	if closeErr := gz.Close(); closeErr != nil {
		_ = f.Close()
		fmt.Printf("Error closing gzip: %v\n", closeErr)
		os.Exit(1)
	}
	if closeErr := f.Close(); closeErr != nil {
		fmt.Printf("Error closing file: %v\n", closeErr)
		os.Exit(1)
	}

	fmt.Println("\nSuccess!")
	fmt.Printf("  - %d samples\n", len(profile.Sample))
	fmt.Printf("  - %d locations\n", len(profile.Location))
	fmt.Printf("  - %d functions\n", len(profile.Function))
	fmt.Printf("  - %d strings\n", len(profile.StringTable))
}

func analyzeCommand(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	topN := fs.Int("top", 20, "Number of top operations to display")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: torch2pprof analyze [options] <input.json>\n")
		fmt.Fprintf(os.Stderr, "\nAnalyze PyTorch profiler trace and show statistics\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(1)
	}

	inputFile := fs.Arg(0)

	traceData, err := converter.LoadTraceFile(inputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	analysis := converter.AnalyzeTrace(traceData)

	fmt.Printf("PyTorch Profile Analysis\n")
	fmt.Printf("========================\n\n")
	fmt.Printf("Total events:           %d\n", analysis.TotalEvents)
	fmt.Printf("Complete events (ph=X): %d\n", analysis.CompleteEvents)
	fmt.Printf("Skipped (dur<=0):       %d\n", analysis.SkippedZeroDuration)
	fmt.Printf("Converted events:       %d\n", analysis.ConvertedEvents)
	fmt.Printf("Unique operations:      %d\n", analysis.UniqueOperations)
	fmt.Printf("Total time:             %.3f ms (%.3f s)\n\n", float64(analysis.TotalTimeNs)/1e6, float64(analysis.TotalTimeNs)/1e9)

	// Display categories
	fmt.Printf("By Category:\n")
	fmt.Printf("%-30s %12s %10s\n", "Category", "Time (ms)", "Count")
	fmt.Printf("%s\n", "------------------------------------------------------")
	for _, c := range analysis.GetSortedCategories() {
		fmt.Printf("%-30s %12.3f %10d\n", c.Name, float64(c.TimeNs)/1e6, c.Count)
	}

	// Top operations
	fmt.Printf("\nTop %d Operations:\n", *topN)
	fmt.Printf("%-60s %12s %10s\n", "Operation", "Time (ms)", "Count")
	fmt.Printf("%s\n", "------------------------------------------------------------------------------------")
	for i, o := range analysis.GetSortedOperations() {
		if i >= *topN {
			break
		}
		name := o.Name
		if len(name) > 58 {
			name = name[:55] + "..."
		}
		fmt.Printf("%-60s %12.3f %10d\n", name, float64(o.TimeNs)/1e6, o.Count)
	}
}
