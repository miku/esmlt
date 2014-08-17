package dupsquash

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/belogik/goes"
)

const AppVersion = "0.1.0"

// SearchConnection is an interface that is satisfied by goes.Connection and
// can be satisfied by mock index connections for testing
type SearchConnection interface {
	Search(query map[string]interface{}, indexList []string, typeList []string, extraArgs url.Values) (goes.Response, error)
}

// ParseIndices parses strings like `2,4,5` into an int slice and adds a `shift`
func ParseIndicesShift(s string, shift int) ([]int, error) {
	parts := strings.Split(s, ",")
	var indices []int
	for _, p := range parts {
		i, err := strconv.ParseInt(p, 10, 0)
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

// ConcatenateValuesNull extracts values according to indices slice and concatenates them
// values that equal a given `nullValue` are ignored
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
	return buffer.String(), nil
}

// ConcatenateValuesNull extracts values according to indices slice and concatenates them,
// uses a default `nullValue` of `<NULL>`
func ConcatenateValues(values []string, indices []int) (string, error) {
	return ConcatenateValuesNull(values, indices, "<NULL>")
}
