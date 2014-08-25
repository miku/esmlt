package esmlt

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/belogik/goes"
)

// Application Version
const Version = "0.1.0"

// SearchConnection is an interface that is satisfied by goes.Connection and
// can be satisfied by mock index connections for testing
type SearchConnection interface {
	Search(query map[string]interface{}, indexList []string, typeList []string, extraArgs url.Values) (goes.Response, error)
}

// ParseIndicesShift parses strings like `2,4,5` into an int slice and adds a `shift`
func ParseIndicesShift(s string, shift int) ([]int, error) {
	parts := strings.Split(s, ",")
	var indices []int
	for _, p := range parts {
		value := strings.TrimSpace(p)
		if value == "" {
			continue
		}
		i, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return nil, err
		}
		indices = append(indices, int(i)+shift)
	}
	return indices, nil
}

// ParseIndices parses strings like `2,4,5` into an int slice
func ParseIndices(s string) ([]int, error) {
	return ParseIndicesShift(s, -1)
}

// ConcatenateValuesNull extracts values according to indices slice and concatenates them.
// Values that equal a given `nullValue` are ignored.
func ConcatenateValuesNull(values []string, indices []int, nullValue string) (string, error) {
	var buffer bytes.Buffer
	for _, i := range indices {
		if i > len(values)-1 {
			return "", fmt.Errorf("index %d exceeds array", i)
		}
		if values[i] == nullValue {
			buffer.WriteString("")
			continue
		}
		buffer.WriteString(values[i])
		buffer.WriteString(" ")
	}
	return strings.TrimSpace(buffer.String()), nil
}

// ConcatenateValues extracts values according to indices slice and concatenates them,
// uses a default `nullValue` of `<NULL>`
func ConcatenateValues(values []string, indices []int) (string, error) {
	return ConcatenateValuesNull(values, indices, "<NULL>")
}

// Value returns the value in a (nested) map according to a key in dot notation.
// If the value is a slice, only the first element is considered.
func Value(key string, doc map[string]interface{}) interface{} {
	keys := strings.Split(key, ".")
	for _, k := range keys {
		value := doc[k]
		if value == nil {
			return nil
		}
		switch value.(type) {
		case map[string]interface{}:
			if len(keys[1:]) == 0 {
				return nil
			}
			return Value(strings.Join(keys[1:], "."), value.(map[string]interface{}))
		case []interface{}:
			if len(value.([]interface{})) == 0 {
				return nil
			}
			first := value.([]interface{})[0]
			switch first.(type) {
			case map[string]interface{}:
				if len(keys[1:]) == 0 {
					return nil
				}
				return Value(strings.Join(keys[1:], "."), first.(map[string]interface{}))
			case string:
				return first
				continue
			}
		default:
			return value
		}
	}
	return nil
}
