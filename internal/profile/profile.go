package profile

import (
	"sync"
)

// ValueType represents the value type for samples in a pprof profile
type ValueType struct {
	Type int64
	Unit int64
}

// Sample represents a single sample in a pprof profile
type Sample struct {
	LocationId []uint64
	Value      []int64
}

// Line represents a line of code in a function
type Line struct {
	FunctionId uint64
	Line       int64
}

// Location represents a location (line of code) in the profile
type Location struct {
	Id   uint64
	Line []*Line
}

// Function represents a function in the profile
type Function struct {
	Id         uint64
	Name       int64
	SystemName int64
	Filename   int64
}

// Profile represents a pprof profile
type Profile struct {
	SampleType    []*ValueType
	Sample        []*Sample
	Location      []*Location
	Function      []*Function
	StringTable   []string
	TimeNanos     int64
	DurationNanos int64
	PeriodType    *ValueType
	Period        int64
}

// Encode encodes the profile to protobuf format
func (p *Profile) Encode() ([]byte, error) {
	var buf []byte

	for _, st := range p.SampleType {
		msg := encodeValueType(st)
		buf = append(buf, encodeTag(1, 2)...)
		buf = append(buf, encodeVarint(uint64(len(msg)))...)
		buf = append(buf, msg...)
	}

	for _, s := range p.Sample {
		msg := encodeSample(s)
		buf = append(buf, encodeTag(2, 2)...)
		buf = append(buf, encodeVarint(uint64(len(msg)))...)
		buf = append(buf, msg...)
	}

	for _, loc := range p.Location {
		msg := encodeLocation(loc)
		buf = append(buf, encodeTag(4, 2)...)
		buf = append(buf, encodeVarint(uint64(len(msg)))...)
		buf = append(buf, msg...)
	}

	for _, fn := range p.Function {
		msg := encodeFunction(fn)
		buf = append(buf, encodeTag(5, 2)...)
		buf = append(buf, encodeVarint(uint64(len(msg)))...)
		buf = append(buf, msg...)
	}

	for _, s := range p.StringTable {
		buf = append(buf, encodeTag(6, 2)...)
		strBytes := []byte(s)
		buf = append(buf, encodeVarint(uint64(len(strBytes)))...)
		buf = append(buf, strBytes...)
	}

	if p.TimeNanos != 0 {
		buf = append(buf, encodeTag(9, 0)...)
		buf = append(buf, encodeVarint(uint64(p.TimeNanos))...)
	}

	if p.DurationNanos != 0 {
		buf = append(buf, encodeTag(10, 0)...)
		buf = append(buf, encodeVarint(uint64(p.DurationNanos))...)
	}

	if p.PeriodType != nil {
		msg := encodeValueType(p.PeriodType)
		buf = append(buf, encodeTag(11, 2)...)
		buf = append(buf, encodeVarint(uint64(len(msg)))...)
		buf = append(buf, msg...)
	}

	if p.Period != 0 {
		buf = append(buf, encodeTag(12, 0)...)
		buf = append(buf, encodeVarint(uint64(p.Period))...)
	}

	return buf, nil
}

func encodeTag(fieldNum int, wireType int) []byte {
	return encodeVarint(uint64((fieldNum << 3) | wireType))
}

func encodeVarint(v uint64) []byte {
	var buf []byte
	for v >= 0x80 {
		buf = append(buf, byte(v)|0x80)
		v >>= 7
	}
	buf = append(buf, byte(v))
	return buf
}

func encodeValueType(vt *ValueType) []byte {
	var buf []byte
	buf = append(buf, encodeTag(1, 0)...)
	buf = append(buf, encodeVarint(uint64(vt.Type))...)
	buf = append(buf, encodeTag(2, 0)...)
	buf = append(buf, encodeVarint(uint64(vt.Unit))...)
	return buf
}

func encodeSample(s *Sample) []byte {
	var buf []byte
	if len(s.LocationId) > 0 {
		var packed []byte
		for _, id := range s.LocationId {
			packed = append(packed, encodeVarint(id)...)
		}
		buf = append(buf, encodeTag(1, 2)...)
		buf = append(buf, encodeVarint(uint64(len(packed)))...)
		buf = append(buf, packed...)
	}
	if len(s.Value) > 0 {
		var packed []byte
		for _, v := range s.Value {
			packed = append(packed, encodeVarint(uint64(v))...)
		}
		buf = append(buf, encodeTag(2, 2)...)
		buf = append(buf, encodeVarint(uint64(len(packed)))...)
		buf = append(buf, packed...)
	}
	return buf
}

func encodeLocation(loc *Location) []byte {
	var buf []byte
	buf = append(buf, encodeTag(1, 0)...)
	buf = append(buf, encodeVarint(loc.Id)...)
	for _, line := range loc.Line {
		msg := encodeLine(line)
		buf = append(buf, encodeTag(4, 2)...)
		buf = append(buf, encodeVarint(uint64(len(msg)))...)
		buf = append(buf, msg...)
	}
	return buf
}

func encodeLine(line *Line) []byte {
	var buf []byte
	buf = append(buf, encodeTag(1, 0)...)
	buf = append(buf, encodeVarint(line.FunctionId)...)
	if line.Line != 0 {
		buf = append(buf, encodeTag(2, 0)...)
		buf = append(buf, encodeVarint(uint64(line.Line))...)
	}
	return buf
}

func encodeFunction(fn *Function) []byte {
	var buf []byte
	buf = append(buf, encodeTag(1, 0)...)
	buf = append(buf, encodeVarint(fn.Id)...)
	buf = append(buf, encodeTag(2, 0)...)
	buf = append(buf, encodeVarint(uint64(fn.Name))...)
	buf = append(buf, encodeTag(3, 0)...)
	buf = append(buf, encodeVarint(uint64(fn.SystemName))...)
	buf = append(buf, encodeTag(4, 0)...)
	buf = append(buf, encodeVarint(uint64(fn.Filename))...)
	return buf
}

// Builder provides thread-safe profile construction
type Builder struct {
	profile       *Profile
	stringIndex   map[string]int64
	functionIndex map[string]uint64
	locationIndex map[string]uint64
	mu            sync.RWMutex
}

// NewBuilder creates a new profile builder
func NewBuilder() *Builder {
	pb := &Builder{
		profile: &Profile{
			StringTable: []string{""},
		},
		stringIndex:   map[string]int64{"": 0},
		functionIndex: map[string]uint64{},
		locationIndex: map[string]uint64{},
	}
	return pb
}

// AddString adds a string to the string table and returns its index
func (pb *Builder) AddString(s string) int64 {
	pb.mu.RLock()
	if idx, ok := pb.stringIndex[s]; ok {
		pb.mu.RUnlock()
		return idx
	}
	pb.mu.RUnlock()

	pb.mu.Lock()
	defer pb.mu.Unlock()
	// Double-check after acquiring write lock
	if idx, ok := pb.stringIndex[s]; ok {
		return idx
	}
	idx := int64(len(pb.profile.StringTable))
	pb.profile.StringTable = append(pb.profile.StringTable, s)
	pb.stringIndex[s] = idx
	return idx
}

// GetOrCreateFunction gets or creates a function and returns its ID
func (pb *Builder) GetOrCreateFunction(name, filename string) uint64 {
	key := name + "\x00" + filename

	pb.mu.RLock()
	if id, ok := pb.functionIndex[key]; ok {
		pb.mu.RUnlock()
		return id
	}
	pb.mu.RUnlock()

	pb.mu.Lock()
	defer pb.mu.Unlock()
	if id, ok := pb.functionIndex[key]; ok {
		return id
	}
	id := uint64(len(pb.profile.Function) + 1)
	fn := &Function{
		Id:         id,
		Name:       pb.addStringLocked(name),
		SystemName: pb.addStringLocked(name),
		Filename:   pb.addStringLocked(filename),
	}
	pb.profile.Function = append(pb.profile.Function, fn)
	pb.functionIndex[key] = id
	return id
}

func (pb *Builder) addStringLocked(s string) int64 {
	if idx, ok := pb.stringIndex[s]; ok {
		return idx
	}
	idx := int64(len(pb.profile.StringTable))
	pb.profile.StringTable = append(pb.profile.StringTable, s)
	pb.stringIndex[s] = idx
	return idx
}

// GetOrCreateLocation gets or creates a location and returns its ID
func (pb *Builder) GetOrCreateLocation(name, filename string) uint64 {
	key := name + "\x00" + filename

	pb.mu.RLock()
	if id, ok := pb.locationIndex[key]; ok {
		pb.mu.RUnlock()
		return id
	}
	pb.mu.RUnlock()

	pb.mu.Lock()
	defer pb.mu.Unlock()
	if id, ok := pb.locationIndex[key]; ok {
		return id
	}

	// Get or create function while holding lock
	funcId := pb.functionIndex[key]
	if funcId == 0 {
		funcId = uint64(len(pb.profile.Function) + 1)
		fn := &Function{
			Id:         funcId,
			Name:       pb.addStringLocked(name),
			SystemName: pb.addStringLocked(name),
			Filename:   pb.addStringLocked(filename),
		}
		pb.profile.Function = append(pb.profile.Function, fn)
		pb.functionIndex[key] = funcId
	}

	id := uint64(len(pb.profile.Location) + 1)
	loc := &Location{
		Id:   id,
		Line: []*Line{{FunctionId: funcId}},
	}
	pb.profile.Location = append(pb.profile.Location, loc)
	pb.locationIndex[key] = id
	return id
}

// SetSampleTypes sets the sample types in the profile
func (pb *Builder) SetSampleTypes(types []struct{ Type, Unit string }) {
	for _, t := range types {
		pb.profile.SampleType = append(pb.profile.SampleType, &ValueType{
			Type: pb.AddString(t.Type),
			Unit: pb.AddString(t.Unit),
		})
	}
}

// SetPeriodType sets the period type in the profile
func (pb *Builder) SetPeriodType(typeName, unit string) {
	pb.profile.PeriodType = &ValueType{
		Type: pb.AddString(typeName),
		Unit: pb.AddString(unit),
	}
}

// Build returns the constructed profile
func (pb *Builder) Build() *Profile {
	return pb.profile
}
