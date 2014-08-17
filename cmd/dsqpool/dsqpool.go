package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/belogik/goes"
	flags "github.com/jessevdk/go-flags"
	"github.com/miku/dupsquash"
)

// Query runs `query` over connection `conn` on `indices` and returns a slice of string slices
func Query(conn *goes.Connection, indices *[]string, query *map[string]interface{}) [][]string {
	extraArgs := make(url.Values, 1)
	searchResults, err := conn.Search(*query, *indices, []string{""}, extraArgs)
	if err != nil {
		log.Fatalln(err)
	}
	results := make([][]string, len(searchResults.Hits.Hits))
	for i, hit := range searchResults.Hits.Hits {
		results[i] = []string{hit.Index, hit.Id, strconv.FormatFloat(hit.Score, 'f', 3, 64)}
	}
	return results
}

// ParseIndices parses strings like `2,4,5` into an int slice
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

func main() {
	var opts struct {
		ElasticSearchHost string `long:"host" default:"localhost" description:"elasticsearch host" value-name:"HOST"`
		ElasticSearchPort string `long:"port" default:"9200" description:"elasticsearch port" value-name:"PORT"`

		Like string `short:"l" long:"like" description:"string to compare" value-name:"STRING"`

		LikeFile      string `short:"i" long:"file" description:"input file (TSV) with strings to compare" value-name:"FILENAME"`
		FileColumn    string `short:"f" long:"column" default:"1" description:"which column(s) to pick for the comparison" value-name:"COLUMN[S]"`
		FileDelimiter string `long:"delimiter" default:"\t" description:"column delimiter of the file" value-name:"DELIM"`
		FileNullValue string `long:"null-value" default:"<NULL>" description:"value that indicates empty value in input file" value-name:"STRING"`

		Index         string `long:"index" description:"index or indices (space separated)" value-name:"NAME[S]"`
		IndexFields   string `short:"x" default:"content.245.a content.245.b" long:"index-fields"description:"which index fields to use for comparison" value-name:"NAME[S]"`
		MinTermFreq   int    `long:"min-term-freq" description:"passed on lucene option" default:"1" value-name:"N"`
		MaxQueryTerms int    `long:"max-query-terms" description:"passed on lucene option" default:"25" value-name:"N"`
		Size          int    `short:"s" long:"size" description:"number of results per query" default:"5" value-name:"N"`

		ShowVersion bool `short:"V" default:"false" long:"version" description:"show version and exit"`
		ShowHelp    bool `short:"h" default:"false" long:"help" description:"show this help message"`
	}
	argparser := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash)
	_, err := argparser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	if opts.ShowVersion {
		fmt.Printf("%s\n", dupsquash.AppVersion)
		return
	}

	argparser.Usage = fmt.Sprintf("[OPTIONS]")
	if opts.ShowHelp {
		argparser.WriteHelp(os.Stdout)
		return
	}

	conn := goes.NewConnection(opts.ElasticSearchHost, opts.ElasticSearchPort)
	fields := strings.Fields(opts.IndexFields)
	indices := strings.Fields(opts.Index)

	if opts.LikeFile != "" {
		if _, err := os.Stat(opts.LikeFile); os.IsNotExist(err) {
			log.Fatalf("no such file or directory: %s\n", opts.LikeFile)
		}

		file, err := os.Open(opts.LikeFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		projector, err := ParseIndices(opts.FileColumn)
		if err != nil {
			log.Fatalf("could not parse column indices: %s\n", opts.FileColumn)
		}

		for scanner.Scan() {
			values := strings.Split(scanner.Text(), opts.FileDelimiter)
			likeText, err := ConcatenateValuesNull(values, projector, opts.FileNullValue)
			if err != nil {
				log.Fatal(err)
			}

			var query = map[string]interface{}{
				"query": map[string]interface{}{
					"more_like_this": map[string]interface{}{
						"fields":          fields,
						"like_text":       likeText,
						"min_term_freq":   opts.MinTermFreq,
						"max_query_terms": opts.MaxQueryTerms,
					},
				},
				"size": opts.Size,
			}

			results := Query(conn, &indices, &query)
			for _, result := range results {
				parts := append(values, result...)
				fmt.Println(strings.Join(parts, "\t"))
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if opts.Like != "" {
		var query = map[string]interface{}{
			"query": map[string]interface{}{
				"more_like_this": map[string]interface{}{
					"fields":          fields,
					"like_text":       opts.Like,
					"min_term_freq":   opts.MinTermFreq,
					"max_query_terms": opts.MaxQueryTerms,
				},
			},
			"size": opts.Size,
		}

		results := Query(conn, &indices, &query)
		for _, result := range results {
			fmt.Println(strings.Join(result, "\t"))
		}
		return
	}
}
