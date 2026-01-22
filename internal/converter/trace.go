package converter

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"pytorch-to-pprof/internal/profile"
)

// TraceEvent represents a single event in the PyTorch trace
type TraceEvent struct {
	Ph   string      `json:"ph"`
	Cat  string      `json:"cat"`
	Name string      `json:"name"`
	Pid  interface{} `json:"pid"`
	Tid  interface{} `json:"tid"`
	Ts   float64     `json:"ts"`
	Dur  float64     `json:"dur"`
}

// TraceData represents the parsed trace JSON structure
type TraceData struct {
	TraceEvents []TraceEvent `json:"traceEvents"`
}

// eventWithEnd is an internal helper that adds the end time
type eventWithEnd struct {
	TraceEvent
	End float64
}

// stackSample represents an aggregated stack sample
type stackSample struct {
	stack  []string // Stack as strings for aggregation key
	names  []string // Function names
	cats   []string // Categories
	timeNs int64
}

// LoadTraceFile loads and parses a PyTorch trace JSON file.
// Supports both plain JSON and gzip-compressed JSON files.
// Automatically detects compression based on file extension (.gz) or content.
func LoadTraceFile(path string) (*TraceData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file

	// Check if file is gzip compressed by extension or magic number
	isGzip := false
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".gz" {
		isGzip = true
	} else {
		// Check magic number (gzip files start with 0x1f 0x8b)
		header := make([]byte, 2)
		n, err := file.Read(header)
		if err == nil && n == 2 && header[0] == 0x1f && header[1] == 0x8b {
			isGzip = true
		}
		// Reset file position
		file.Seek(0, 0)
	}

	// Wrap with gzip reader if compressed
	if isGzip {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read and parse JSON
	var traceData TraceData
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&traceData); err != nil {
		return nil, err
	}

	return &traceData, nil
}

// getTid converts a tid field to int64
func getTid(tid interface{}) int64 {
	switch v := tid.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case string:
		var h int64
		for _, c := range v {
			h = h*31 + int64(c)
		}
		return h
	default:
		return 0
	}
}

// ProcessThreadEvents processes a single thread's events using a stack-based algorithm.
// This is O(n) instead of O(n²) when compared to naive pairwise comparison.
func ProcessThreadEvents(events []eventWithEnd, pb *profile.Builder, results chan<- stackSample, counter *int64) {
	type stackEntry struct {
		event eventWithEnd
		name  string
		cat   string
	}
	var stack []stackEntry

	for _, event := range events {
		// Pop events from stack that have ended before current event starts
		for len(stack) > 0 && stack[len(stack)-1].event.End < event.Ts {
			stack = stack[:len(stack)-1]
		}

		// Also pop events that end before our event ends (they can't be our parent)
		// Keep only events that fully contain us
		newStack := stack[:0]
		for _, s := range stack {
			if s.event.End >= event.End {
				newStack = append(newStack, s)
			}
		}
		stack = newStack

		// Current stack + this event forms our call stack
		names := make([]string, len(stack)+1)
		cats := make([]string, len(stack)+1)
		stackKey := make([]string, len(stack)+1)

		for i, s := range stack {
			names[i] = s.name
			cats[i] = s.cat
			stackKey[i] = s.name + "\x00" + s.cat
		}
		names[len(stack)] = event.Name
		cats[len(stack)] = event.Cat
		stackKey[len(stack)] = event.Name + "\x00" + event.Cat

		// Push current event to stack
		stack = append(stack, stackEntry{
			event: event,
			name:  event.Name,
			cat:   event.Cat,
		})

		durNs := int64(event.Dur * 1000)

		results <- stackSample{
			stack:  stackKey,
			names:  names,
			cats:   cats,
			timeNs: durNs,
		}

		atomic.AddInt64(counter, 1)
	}
}

// ConvertOptions contains options for trace conversion
type ConvertOptions struct {
	NumWorkers int
}

// sampleData represents aggregated sample data
type sampleData struct {
	locationIds []uint64
	count       int64
	timeNs      int64
}

// ConvertTrace converts PyTorch trace data to a pprof profile
func ConvertTrace(traceData *TraceData, opts ConvertOptions) *profile.Profile {
	// Group events by thread
	threadEvents := make(map[int64][]eventWithEnd)
	for _, e := range traceData.TraceEvents {
		if e.Ph != "X" || e.Dur <= 0 {
			continue
		}
		tid := getTid(e.Tid)
		threadEvents[tid] = append(threadEvents[tid], eventWithEnd{
			TraceEvent: e,
			End:        e.Ts + e.Dur,
		})
	}

	// Sort each thread's events by start time
	for tid := range threadEvents {
		events := threadEvents[tid]
		sort.Slice(events, func(i, j int) bool {
			return events[i].Ts < events[j].Ts
		})
	}

	pb := profile.NewBuilder()
	pb.SetSampleTypes([]struct{ Type, Unit string }{
		{"samples", "count"},
		{"time", "nanoseconds"},
	})
	pb.SetPeriodType("cpu", "nanoseconds")
	pb.Build().Period = 1000000

	// Channel for collecting results from workers
	results := make(chan stackSample, 10000)

	// Progress counter
	var processedCount int64

	// Process threads in parallel
	var wg sync.WaitGroup
	for _, events := range threadEvents {
		wg.Add(1)
		go func(events []eventWithEnd) {
			defer wg.Done()
			ProcessThreadEvents(events, pb, results, &processedCount)
		}(events)
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate results
	sampleMap := make(map[string]*sampleData)

	for sample := range results {
		// Build key from stack
		key := ""
		for _, s := range sample.stack {
			key += s + ";"
		}

		if existing, ok := sampleMap[key]; ok {
			existing.count++
			existing.timeNs += sample.timeNs
		} else {
			// Build location IDs (pprof wants leaf first)
			locationIds := make([]uint64, len(sample.names))
			for i := range sample.names {
				locId := pb.GetOrCreateLocation(sample.names[i], sample.cats[i])
				// Reverse order: leaf first
				locationIds[len(sample.names)-1-i] = locId
			}
			sampleMap[key] = &sampleData{
				locationIds: locationIds,
				count:       1,
				timeNs:      sample.timeNs,
			}
		}
	}

	// Add samples to profile
	for _, s := range sampleMap {
		pb.Build().Sample = append(pb.Build().Sample, &profile.Sample{
			LocationId: s.locationIds,
			Value:      []int64{s.count, s.timeNs},
		})
	}

	return pb.Build()
}
