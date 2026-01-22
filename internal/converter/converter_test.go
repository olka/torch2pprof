package converter

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetTid(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"float64", float64(123.0), 123},
		{"int", 456, 456},
		{"int64", int64(789), 789},
		{"string", "test", int64('t')*31*31*31 + int64('e')*31*31 + int64('s')*31 + int64('t')},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTid(tt.input)
			if result != tt.expected {
				t.Errorf("getTid(%v): expected %d, got %d", tt.input, tt.expected, result)
			}
		})
	}
}

func TestLoadTraceFile_PlainJSON(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "test", Cat: "test_cat", Ts: 100, Dur: 50},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(testFile, data, 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test loading
	loaded, err := LoadTraceFile(testFile)
	if err != nil {
		t.Fatalf("LoadTraceFile failed: %v", err)
	}

	if len(loaded.TraceEvents) != 1 {
		t.Errorf("Expected 1 event, got %d", len(loaded.TraceEvents))
	}

	event := loaded.TraceEvents[0]
	if event.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", event.Name)
	}
	if event.Ts != 100 {
		t.Errorf("Expected Ts 100, got %f", event.Ts)
	}
	if event.Dur != 50 {
		t.Errorf("Expected Dur 50, got %f", event.Dur)
	}
}

func TestLoadTraceFile_GzipJSON(t *testing.T) {
	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json.gz")

	testData := TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "test", Cat: "test_cat", Ts: 100, Dur: 50},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Write gzipped file
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer func() { _ = f.Close() }()

	gz := gzip.NewWriter(f)
	if _, err := gz.Write(data); err != nil {
		t.Fatalf("Failed to write gzip data: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	// Test loading
	loaded, err := LoadTraceFile(testFile)
	if err != nil {
		t.Fatalf("LoadTraceFile failed: %v", err)
	}

	if len(loaded.TraceEvents) != 1 {
		t.Errorf("Expected 1 event, got %d", len(loaded.TraceEvents))
	}

	event := loaded.TraceEvents[0]
	if event.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", event.Name)
	}
}

func TestLoadTraceFile_GzipWithoutExtension(t *testing.T) {
	// Create temporary gzipped file without .gz extension
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_no_ext")

	testData := TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "test", Cat: "test_cat", Ts: 100, Dur: 50},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Write gzipped file
	f, err := os.Create(testFile)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer func() { _ = f.Close() }()

	gz := gzip.NewWriter(f)
	if _, err := gz.Write(data); err != nil {
		t.Fatalf("Failed to write gzip data: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	// Test loading (should detect gzip by magic number)
	loaded, err := LoadTraceFile(testFile)
	if err != nil {
		t.Fatalf("LoadTraceFile failed: %v", err)
	}

	if len(loaded.TraceEvents) != 1 {
		t.Errorf("Expected 1 event, got %d", len(loaded.TraceEvents))
	}
}

func TestLoadTraceFile_NonexistentFile(t *testing.T) {
	_, err := LoadTraceFile("/nonexistent/file.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestAnalyzeTrace(t *testing.T) {
	testData := &TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "op1", Cat: "cat1", Ts: 100, Dur: 50},
			{Ph: "X", Name: "op2", Cat: "cat1", Ts: 200, Dur: 30},
			{Ph: "X", Name: "op1", Cat: "cat2", Ts: 300, Dur: 20},
			{Ph: "B", Name: "op3", Cat: "cat1", Ts: 400, Dur: 10}, // Not "X", should be skipped
			{Ph: "X", Name: "op4", Cat: "cat1", Ts: 500, Dur: 0},  // Zero duration, should be skipped
		},
	}

	analysis := AnalyzeTrace(testData)

	// Check total events
	if analysis.TotalEvents != 5 {
		t.Errorf("Expected 5 total events, got %d", analysis.TotalEvents)
	}

	// Check complete events (Ph == "X")
	if analysis.CompleteEvents != 4 {
		t.Errorf("Expected 4 complete events, got %d", analysis.CompleteEvents)
	}

	// Check skipped zero duration
	if analysis.SkippedZeroDuration != 1 {
		t.Errorf("Expected 1 skipped zero duration, got %d", analysis.SkippedZeroDuration)
	}

	// Check converted events
	if analysis.ConvertedEvents != 3 {
		t.Errorf("Expected 3 converted events, got %d", analysis.ConvertedEvents)
	}

	// Check unique operations
	if analysis.UniqueOperations != 2 { // op1 (appears twice with different cats), op2
		t.Errorf("Expected 2 unique operations, got %d", analysis.UniqueOperations)
	}

	// Check total time (50 + 30 + 20 = 100 microseconds = 100000 nanoseconds)
	expectedTimeNs := int64(100000)
	if analysis.TotalTimeNs != expectedTimeNs {
		t.Errorf("Expected total time %d ns, got %d ns", expectedTimeNs, analysis.TotalTimeNs)
	}

	// Check category stats
	if len(analysis.CategoryStats) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(analysis.CategoryStats))
	}

	cat1Stats := analysis.CategoryStats["cat1"]
	if cat1Stats.Count != 2 {
		t.Errorf("Expected cat1 count 2, got %d", cat1Stats.Count)
	}
	expectedCat1Time := int64(80000) // 50 + 30 microseconds
	if cat1Stats.TimeNs != expectedCat1Time {
		t.Errorf("Expected cat1 time %d ns, got %d ns", expectedCat1Time, cat1Stats.TimeNs)
	}
}

func TestGetSortedCategories(t *testing.T) {
	testData := &TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "op1", Cat: "cat1", Ts: 100, Dur: 10},
			{Ph: "X", Name: "op2", Cat: "cat2", Ts: 200, Dur: 50},
			{Ph: "X", Name: "op3", Cat: "cat3", Ts: 300, Dur: 30},
		},
	}

	analysis := AnalyzeTrace(testData)
	sorted := analysis.GetSortedCategories()

	if len(sorted) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(sorted))
	}

	// Should be sorted by time descending
	if sorted[0].Name != "cat2" { // 50 microseconds
		t.Errorf("Expected first category 'cat2', got '%s'", sorted[0].Name)
	}
	if sorted[1].Name != "cat3" { // 30 microseconds
		t.Errorf("Expected second category 'cat3', got '%s'", sorted[1].Name)
	}
	if sorted[2].Name != "cat1" { // 10 microseconds
		t.Errorf("Expected third category 'cat1', got '%s'", sorted[2].Name)
	}
}

func TestGetSortedOperations(t *testing.T) {
	testData := &TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "fast", Cat: "cat1", Ts: 100, Dur: 10},
			{Ph: "X", Name: "slow", Cat: "cat1", Ts: 200, Dur: 50},
			{Ph: "X", Name: "medium", Cat: "cat1", Ts: 300, Dur: 30},
		},
	}

	analysis := AnalyzeTrace(testData)
	sorted := analysis.GetSortedOperations()

	if len(sorted) != 3 {
		t.Errorf("Expected 3 operations, got %d", len(sorted))
	}

	// Should be sorted by time descending
	if sorted[0].Name != "slow" {
		t.Errorf("Expected first operation 'slow', got '%s'", sorted[0].Name)
	}
	if sorted[1].Name != "medium" {
		t.Errorf("Expected second operation 'medium', got '%s'", sorted[1].Name)
	}
	if sorted[2].Name != "fast" {
		t.Errorf("Expected third operation 'fast', got '%s'", sorted[2].Name)
	}
}

func TestConvertTrace(t *testing.T) {
	testData := &TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "X", Name: "op1", Cat: "cat1", Tid: 1, Ts: 100, Dur: 50},
			{Ph: "X", Name: "op2", Cat: "cat1", Tid: 1, Ts: 110, Dur: 30}, // Nested
			{Ph: "X", Name: "op3", Cat: "cat2", Tid: 2, Ts: 200, Dur: 20},
		},
	}

	profile := ConvertTrace(testData, ConvertOptions{NumWorkers: 2})

	if profile == nil {
		t.Fatal("ConvertTrace returned nil")
	}

	if len(profile.Sample) == 0 {
		t.Error("Expected samples to be created")
	}

	if len(profile.Location) == 0 {
		t.Error("Expected locations to be created")
	}

	if len(profile.Function) == 0 {
		t.Error("Expected functions to be created")
	}

	if len(profile.StringTable) == 0 {
		t.Error("Expected strings to be created")
	}

	// Verify sample types were set
	if len(profile.SampleType) != 2 {
		t.Errorf("Expected 2 sample types, got %d", len(profile.SampleType))
	}
}

func TestConvertTrace_EmptyEvents(t *testing.T) {
	testData := &TraceData{
		TraceEvents: []TraceEvent{},
	}

	profile := ConvertTrace(testData, ConvertOptions{NumWorkers: 1})

	if profile == nil {
		t.Fatal("ConvertTrace returned nil")
	}

	if len(profile.Sample) != 0 {
		t.Errorf("Expected 0 samples for empty trace, got %d", len(profile.Sample))
	}
}

func TestConvertTrace_FilteredEvents(t *testing.T) {
	testData := &TraceData{
		TraceEvents: []TraceEvent{
			{Ph: "B", Name: "op1", Cat: "cat1", Tid: 1, Ts: 100, Dur: 50}, // Wrong phase
			{Ph: "X", Name: "op2", Cat: "cat1", Tid: 1, Ts: 200, Dur: 0},  // Zero duration
			{Ph: "X", Name: "op3", Cat: "cat1", Tid: 1, Ts: 300, Dur: -1}, // Negative duration
		},
	}

	profile := ConvertTrace(testData, ConvertOptions{NumWorkers: 1})

	if profile == nil {
		t.Fatal("ConvertTrace returned nil")
	}

	// All events should be filtered out
	if len(profile.Sample) != 0 {
		t.Errorf("Expected 0 samples (all filtered), got %d", len(profile.Sample))
	}
}
