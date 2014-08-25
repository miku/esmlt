package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/belogik/goes"
)

type Work struct {
	Indices    []string
	Connection dupsquash.SearchConnection

	NullValue     string
	Fields        []string
	LikeText      string
	MinTermFreq   int
	MaxQueryTerms int
	Size          int

	Values []string
}

func Worker(in chan *Work, out chan [][]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for work := range in {
		var query = map[string]interface{}{
			"query": map[string]interface{}{
				"more_like_this": map[string]interface{}{
					"fields":          work.Fields,
					"like_text":       work.LikeText,
					"min_term_freq":   work.MinTermFreq,
					"max_query_terms": work.MaxQueryTerms,
				},
			},
			"size": work.Size,
		}
		queryResults := QueryField(work, &query)
		var results [][]string
		for _, result := range queryResults {
			parts := append(work.Values, result...)
			results = append(results, parts)
		}
		out <- results
	}
}

// FanInWriter writes the channel content to the writer
func FanInWriter(writer io.Writer, in chan [][]string, done chan bool) {
	for results := range in {
		for _, result := range results {
			writer.Write([]byte(strings.Join(result, "\t")))
			writer.Write([]byte("\n"))
		}
	}
	done <- true
}

func QueryField(work *Work, query *map[string]interface{}) [][]string {
	extraArgs := make(url.Values, 1)
	searchResults, err := work.Connection.Search(*query, work.Indices, []string{""}, extraArgs)
	if err != nil {
		log.Fatalln(err)
	}
	var results [][]string
	for _, hit := range searchResults.Hits.Hits {
		var values []string
		for _, field := range work.Fields {
			value := dupsquash.Value(field, hit.Source)
			switch value.(type) {
			case string:
				values = append(values, value.(string))
			case nil:
				values = append(values, work.NullValue)
			default:
				continue
			}
		}
		fields := append([]string{hit.Index, hit.Type, hit.Id, strconv.FormatFloat(hit.Score, 'f', 3, 64)}, values...)
		results = append(results, fields)
	}
	return results
}

func main() {

	esHost := flag.String("host", "localhost", "elasticsearch host")
	esPort := flag.String("port", "9200", "elasticsearch port")
	likeText := flag.String("like", "", "more like this queries like-text")
	likeFile := flag.String("file", "", "input file")
	fileColumn := flag.String("columns", "1", "which column to use as like-text")
	columnDelimiter := flag.String("delimiter", "\t", "column delimiter of the input file")
	columnNull := flag.String("null", "NOT_AVAILABLE", "column value to ignore")
	indicesString := flag.String("indices", "", "index or indices to query")
	indexFields := flag.String("fields", "content.245.a content.245.b", "index fields to query")
	minTermFreq := flag.Int("min-term-freq", 1, "min term frequency")
	maxQueryTerms := flag.Int("max-query-terms", 25, "max query terms")
	size := flag.Int("size", 5, "maximum number of similar records to report")
	numWorkers := flag.Int("workers", runtime.NumCPU(), "number of workers to use")
	version := flag.Bool("v", false, "prints current program version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if *version {
		fmt.Printf("%s\n", dupsquash.AppVersion)
		return
	}

	if *likeText == "" && *likeFile == "" {
		PrintUsage()
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	runtime.GOMAXPROCS(*numWorkers)

	conn := goes.NewConnection(*esHost, *esPort)
	fields := strings.Fields(*indexFields)
	indices := strings.Fields(*indicesString)

	if *likeFile != "" {
		if _, err := os.Stat(*likeFile); os.IsNotExist(err) {
			log.Fatalf("no such file or directory: %s\n", *likeFile)
		}

		file, err := os.Open(*likeFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		projector, err := dupsquash.ParseIndices(*fileColumn)
		if err != nil {
			log.Fatalf("could not parse column indices: %s\n", *fileColumn)
		}

		queue := make(chan *Work)
		results := make(chan [][]string)
		done := make(chan bool)

		writer := bufio.NewWriter(os.Stdout)
		defer writer.Flush()
		go FanInWriter(writer, results, done)

		var wg sync.WaitGroup
		for i := 0; i < *numWorkers; i++ {
			wg.Add(1)
			go Worker(queue, results, &wg)
		}

		for scanner.Scan() {
			values := strings.Split(scanner.Text(), *columnDelimiter)
			likeText, err := dupsquash.ConcatenateValuesNull(values, projector, *columnNull)
			if err != nil {
				log.Fatal(err)
			}

			work := Work{
				Indices:       indices,
				Connection:    conn,
				Fields:        fields,
				NullValue:     *columnNull,
				LikeText:      likeText,
				MinTermFreq:   *minTermFreq,
				MaxQueryTerms: *maxQueryTerms,
				Size:          *size,
				Values:        values,
			}
			queue <- &work
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		close(queue)
		wg.Wait()
		close(results)
		select {
		case <-time.After(1e9):
			break
		case <-done:
			break
		}
		return
	}

	if *likeText != "" {
		var query = map[string]interface{}{
			"query": map[string]interface{}{
				"more_like_this": map[string]interface{}{
					"fields":          fields,
					"like_text":       *likeText,
					"min_term_freq":   *minTermFreq,
					"max_query_terms": *maxQueryTerms,
				},
			},
			"size": *size,
		}

		work := Work{
			Indices:       indices,
			Connection:    conn,
			Fields:        fields,
			NullValue:     *columnNull,
			LikeText:      *likeText,
			MinTermFreq:   *minTermFreq,
			MaxQueryTerms: *maxQueryTerms,
			Size:          *size,
			Values:        []string{},
		}
		results := QueryField(&work, &query)
		for _, result := range results {
			fmt.Println(strings.Join(result, "\t"))
		}
		return
	}
}
