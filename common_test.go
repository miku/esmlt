package esmlt

import (
	"reflect"
	"testing"
)

func TestParseIndicesShift(t *testing.T) {
	var tests = []struct {
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

	for _, tt := range tests {
		out, err := ParseIndicesShift(tt.s, tt.shift)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("ParseIndicesShift(%s, %d) => %v, %v, want: %v", tt.s, tt.shift, out, err, tt.out)
		}
	}
}

func TestParseIndices(t *testing.T) {
	var tests = []struct {
		s   string
		out []int
	}{
		{"1", []int{0}},
		{"1,2", []int{0, 1}},
	}

	for _, tt := range tests {
		out, err := ParseIndices(tt.s)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("ParseIndices(%s) => %v, %v, want: %v", tt.s, out, err, tt.out)
		}
	}
}

func TestConcatenateValuesNull(t *testing.T) {
	var tests = []struct {
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

	for _, tt := range tests {
		out, err := ConcatenateValuesNull(tt.values, tt.indices, tt.nullValue)
		if err == nil && tt.errNotNil {
			t.Errorf("ConcatenateValuesNull(%v, %v, %s) => expected non-nil err", tt.values, tt.indices, tt.nullValue)
		}
		if out != tt.out {
			t.Errorf("ConcatenateValuesNull(%v, %v, %s) => %v, %v, want: %v", tt.values, tt.indices, tt.nullValue, out, err, tt.out)
		}
	}
}

func TestValue(t *testing.T) {
	var tests = []struct {
		key string
		doc map[string]interface{}
		out interface{}
	}{
		{"a", map[string]interface{}{"a": 1, "b": 2}, 1},
		{"b", map[string]interface{}{"a": 1, "b": 2}, 2},
		{"c", map[string]interface{}{"a": 1, "b": 2}, nil},
		{"a", map[string]interface{}{"a": []int{1, 2}, "b": 2}, []int{1, 2}},
		{"a.c", map[string]interface{}{"a": []interface{}{map[string]interface{}{"c": "3"}, "d"}, "b": 2}, "3"},
		{"a.b", map[string]interface{}{"a": map[string]interface{}{"b": "22"}, "b": 2}, "22"},
		{"a.c", map[string]interface{}{"a": map[string]interface{}{"b": "22"}, "b": 2}, nil},
		{"a.b.c", map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": "ccc"}, "c": 2}}, "ccc"},
	}
	for _, tt := range tests {
		out := Value(tt.key, tt.doc)
		if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("Value(%s, %v) => %v, want: %v", tt.key, tt.doc, out, tt.out)
		}
	}
}
