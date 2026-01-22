package converter

import (
	"sort"
)

// CategoryStats holds statistics for a category
type CategoryStats struct {
	Count  int
	TimeNs int64
}

// OperationStats holds statistics for an operation
type OperationStats struct {
	Count  int
	TimeNs int64
}

// TraceAnalysis contains analysis results from a trace
type TraceAnalysis struct {
	TotalEvents         int
	CompleteEvents      int
	SkippedZeroDuration int
	ConvertedEvents     int
	UniqueOperations    int
	TotalTimeNs         int64
	CategoryStats       map[string]CategoryStats
	OperationStats      map[string]OperationStats
}

// AnalyzeTrace analyzes a PyTorch trace and returns statistics
func AnalyzeTrace(traceData *TraceData) *TraceAnalysis {
	analysis := &TraceAnalysis{
		CategoryStats:  make(map[string]CategoryStats),
		OperationStats: make(map[string]OperationStats),
	}

	for _, e := range traceData.TraceEvents {
		analysis.TotalEvents++
		if e.Ph != "X" {
			continue
		}
		analysis.CompleteEvents++
		if e.Dur <= 0 {
			analysis.SkippedZeroDuration++
			continue
		}

		analysis.ConvertedEvents++
		durNs := int64(e.Dur * 1000)
		analysis.TotalTimeNs += durNs

		// By category
		cs := analysis.CategoryStats[e.Cat]
		cs.Count++
		cs.TimeNs += durNs
		analysis.CategoryStats[e.Cat] = cs

		// By operation
		os := analysis.OperationStats[e.Name]
		os.Count++
		os.TimeNs += durNs
		analysis.OperationStats[e.Name] = os
	}

	analysis.UniqueOperations = len(analysis.OperationStats)

	return analysis
}

// CategoryEntry is a helper for sorting categories
type CategoryEntry struct {
	Name   string
	Count  int
	TimeNs int64
}

// GetSortedCategories returns categories sorted by time descending
func (a *TraceAnalysis) GetSortedCategories() []CategoryEntry {
	entries := make([]CategoryEntry, 0, len(a.CategoryStats))
	for name, s := range a.CategoryStats {
		entries = append(entries, CategoryEntry{name, s.Count, s.TimeNs})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].TimeNs > entries[j].TimeNs })
	return entries
}

// OperationEntry is a helper for sorting operations
type OperationEntry struct {
	Name   string
	Count  int
	TimeNs int64
}

// GetSortedOperations returns operations sorted by time descending
func (a *TraceAnalysis) GetSortedOperations() []OperationEntry {
	entries := make([]OperationEntry, 0, len(a.OperationStats))
	for name, s := range a.OperationStats {
		entries = append(entries, OperationEntry{name, s.Count, s.TimeNs})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].TimeNs > entries[j].TimeNs })
	return entries
}
