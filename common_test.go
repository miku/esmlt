package dupsquash

import (
	"reflect"
	"testing"
)

var parseIndicesShiftTests = []struct {
	s     string
	shift int
	out   []int
}{
	{"1", 0, []int{1}},
	{"1,2", 0, []int{1, 2}},
	{"1,  2,3", 0, []int{1, 2, 3}},
	{"1,  2,,,3", 0, []int{1, 2, 3}},
	{"1a,  2,,,3", 0, nil},
	{"2,3", -1, []int{1, 2}},
	{"2,3", -10, []int{-8, -7}},
	{"10,11,12", -10, []int{0, 1, 2}},
}

func TestParseIndicesShift(t *testing.T) {
	for _, tt := range parseIndicesShiftTests {
		out, err := ParseIndicesShift(tt.s, tt.shift)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("ParseIndicesShift(%s, %d) => %v, %v, want: %v", tt.s, tt.shift, out, err, tt.out)
		}
	}
}

var parseIndicesTests = []struct {
	s   string
	out []int
}{
	{"1", []int{0}},
	{"1,2", []int{0, 1}},
}

func TestParseIndices(t *testing.T) {
	for _, tt := range parseIndicesTests {
		out, err := ParseIndices(tt.s)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("ParseIndices(%s) => %v, %v, want: %v", tt.s, out, err, tt.out)
		}
	}
}

var concatenateValuesNullTests = []struct {
	values    []string
	indices   []int
	nullValue string
	out       string
	errNotNil bool
}{
	{[]string{"A", "B"}, []int{1}, "", "B", false},
	{[]string{"A", "B", "C"}, []int{0, 2}, "", "A C", false},
	{[]string{"A", "B", "X"}, []int{0, 2}, "X", "A", false},
	{[]string{"A", "B", "X"}, []int{10}, "", "", true},
}

func TestConcatenateValuesNull(t *testing.T) {
	for _, tt := range concatenateValuesNullTests {
		out, err := ConcatenateValuesNull(tt.values, tt.indices, tt.nullValue)
		if err == nil && tt.errNotNil {
			t.Errorf("ConcatenateValuesNull(%v, %v, %s) => expected non-nil err", tt.values, tt.indices, tt.nullValue)
		}
		if out != tt.out {
			t.Errorf("ConcatenateValuesNull(%v, %v, %s) => %v, %v, want: %v", tt.values, tt.indices, tt.nullValue, out, err, tt.out)
		}
	}
}
