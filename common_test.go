package dupsquash

import (
	"reflect"
	"testing"
)

var parseIndicesShiftTests = []struct {
	s     string
	shift int
	out   []int
	err   error
}{
	{"1", 0, []int{1}, nil},
}

func TestParseIndicesShift(t *testing.T) {
	for _, tt := range parseIndicesShiftTests {
		out, err := ParseIndicesShift(tt.s, tt.shift)
		if err != tt.err {
			t.Errorf("ParseIndicesShift(%s, %d) => %v, %v, want: %v, %v", tt.s, tt.shift, out, err, tt.out, tt.err)
		}
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("ParseIndicesShift(%s, %d) => %v, %v, want: %v, %v", tt.s, tt.shift, out, err, tt.out, tt.err)
		}
	}
}
