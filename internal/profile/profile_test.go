package profile

import (
	"testing"
)

func TestNewBuilder(t *testing.T) {
	pb := NewBuilder()
	if pb == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if pb.profile == nil {
		t.Fatal("Builder profile is nil")
	}
	if len(pb.profile.StringTable) != 1 || pb.profile.StringTable[0] != "" {
		t.Error("StringTable should start with empty string")
	}
	if pb.stringIndex[""] != 0 {
		t.Error("Empty string should have index 0")
	}
}

func TestAddString(t *testing.T) {
	pb := NewBuilder()

	// Add first string
	idx1 := pb.AddString("test")
	if idx1 != 1 {
		t.Errorf("Expected index 1, got %d", idx1)
	}

	// Add same string again (should return same index)
	idx2 := pb.AddString("test")
	if idx2 != idx1 {
		t.Errorf("Expected same index %d, got %d", idx1, idx2)
	}

	// Add different string
	idx3 := pb.AddString("other")
	if idx3 != 2 {
		t.Errorf("Expected index 2, got %d", idx3)
	}

	// Verify string table
	if len(pb.profile.StringTable) != 3 {
		t.Errorf("Expected 3 strings, got %d", len(pb.profile.StringTable))
	}
}

func TestGetOrCreateFunction(t *testing.T) {
	pb := NewBuilder()

	// Create first function
	id1 := pb.GetOrCreateFunction("func1", "file1.py")
	if id1 != 1 {
		t.Errorf("Expected function ID 1, got %d", id1)
	}

	// Get same function (should return same ID)
	id2 := pb.GetOrCreateFunction("func1", "file1.py")
	if id2 != id1 {
		t.Errorf("Expected same ID %d, got %d", id1, id2)
	}

	// Create different function
	id3 := pb.GetOrCreateFunction("func2", "file2.py")
	if id3 != 2 {
		t.Errorf("Expected function ID 2, got %d", id3)
	}

	// Verify functions created
	if len(pb.profile.Function) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(pb.profile.Function))
	}

	// Verify function details
	fn := pb.profile.Function[0]
	if fn.Id != 1 {
		t.Errorf("Expected function ID 1, got %d", fn.Id)
	}
}

func TestGetOrCreateLocation(t *testing.T) {
	pb := NewBuilder()

	// Create first location
	id1 := pb.GetOrCreateLocation("func1", "file1.py")
	if id1 != 1 {
		t.Errorf("Expected location ID 1, got %d", id1)
	}

	// Get same location (should return same ID)
	id2 := pb.GetOrCreateLocation("func1", "file1.py")
	if id2 != id1 {
		t.Errorf("Expected same ID %d, got %d", id1, id2)
	}

	// Create different location
	id3 := pb.GetOrCreateLocation("func2", "file2.py")
	if id3 != 2 {
		t.Errorf("Expected location ID 2, got %d", id3)
	}

	// Verify locations and functions created
	if len(pb.profile.Location) != 2 {
		t.Errorf("Expected 2 locations, got %d", len(pb.profile.Location))
	}
	if len(pb.profile.Function) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(pb.profile.Function))
	}

	// Verify location has line with function reference
	loc := pb.profile.Location[0]
	if len(loc.Line) != 1 {
		t.Errorf("Expected 1 line, got %d", len(loc.Line))
	}
	if loc.Line[0].FunctionId != 1 {
		t.Errorf("Expected function ID 1, got %d", loc.Line[0].FunctionId)
	}
}

func TestSetSampleTypes(t *testing.T) {
	pb := NewBuilder()

	types := []struct{ Type, Unit string }{
		{"samples", "count"},
		{"time", "nanoseconds"},
	}
	pb.SetSampleTypes(types)

	if len(pb.profile.SampleType) != 2 {
		t.Errorf("Expected 2 sample types, got %d", len(pb.profile.SampleType))
	}

	// Verify strings were added
	if len(pb.profile.StringTable) < 5 { // "", "samples", "count", "time", "nanoseconds"
		t.Errorf("Expected at least 5 strings, got %d", len(pb.profile.StringTable))
	}
}

func TestSetPeriodType(t *testing.T) {
	pb := NewBuilder()

	pb.SetPeriodType("cpu", "nanoseconds")

	if pb.profile.PeriodType == nil {
		t.Fatal("PeriodType is nil")
	}

	// Verify strings were added
	if len(pb.profile.StringTable) < 3 { // "", "cpu", "nanoseconds"
		t.Errorf("Expected at least 3 strings, got %d", len(pb.profile.StringTable))
	}
}

func TestProfileEncode(t *testing.T) {
	pb := NewBuilder()

	// Set up a minimal profile
	pb.SetSampleTypes([]struct{ Type, Unit string }{
		{"samples", "count"},
	})
	pb.SetPeriodType("cpu", "nanoseconds")

	// Create a location
	locId := pb.GetOrCreateLocation("test_func", "test.py")

	// Add a sample
	pb.profile.Sample = append(pb.profile.Sample, &Sample{
		LocationId: []uint64{locId},
		Value:      []int64{1},
	})

	// Build and encode
	profile := pb.Build()
	data, err := profile.Encode()

	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Encoded data is empty")
	}

	// Basic sanity check - encoded data should be substantial
	if len(data) < 20 {
		t.Errorf("Encoded data seems too small: %d bytes", len(data))
	}
}

func TestEncodeVarint(t *testing.T) {
	tests := []struct {
		input    uint64
		expected []byte
	}{
		{0, []byte{0}},
		{1, []byte{1}},
		{127, []byte{127}},
		{128, []byte{0x80, 0x01}},
		{300, []byte{0xac, 0x02}},
	}

	for _, tt := range tests {
		result := encodeVarint(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("encodeVarint(%d): expected length %d, got %d", tt.input, len(tt.expected), len(result))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("encodeVarint(%d): expected %v, got %v", tt.input, tt.expected, result)
				break
			}
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	pb := NewBuilder()

	// Test concurrent string addition
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				pb.AddString("test")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// "test" should only appear once in the string table
	count := 0
	for _, s := range pb.profile.StringTable {
		if s == "test" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("Expected 'test' to appear once, found %d times", count)
	}
}

func TestBuild(t *testing.T) {
	pb := NewBuilder()

	pb.SetSampleTypes([]struct{ Type, Unit string }{
		{"samples", "count"},
		{"time", "nanoseconds"},
	})

	profile := pb.Build()

	if profile == nil {
		t.Fatal("Build returned nil")
	}

	if len(profile.SampleType) != 2 {
		t.Errorf("Expected 2 sample types, got %d", len(profile.SampleType))
	}
}
