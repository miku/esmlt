package main

import (
	"bufio"
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
	flags "github.com/jessevdk/go-flags"
	"github.com/miku/dupsquash"
)

type Work struct {
	Indices    []string
	Connection dupsquash.SearchConnection

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
		queryResults := Query(work.Connection, &work.Indices, &query)
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

// Query runs `query` over connection `conn` on `indices` and returns a slice of string slices
func Query(conn dupsquash.SearchConnection, indices *[]string, query *map[string]interface{}) [][]string {
	extraArgs := make(url.Values, 1)
	searchResults, err := conn.Search(*query, *indices, []string{""}, extraArgs)
	if err != nil {
		log.Fatalln(err)
	}
	results := make([][]string, len(searchResults.Hits.Hits))
	for i, hit := range searchResults.Hits.Hits {
		results[i] = []string{hit.Index, hit.Type, hit.Id, strconv.FormatFloat(hit.Score, 'f', 3, 64)}
	}
	return results
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

		NumWorkers  int    `short:"w" long:"workers" default:"0" description:"number of workers, 0 means number of available cpus" value-name:"N"`
		ShowVersion bool   `short:"V" default:"false" long:"version" description:"show version and exit"`
		ShowHelp    bool   `short:"h" default:"false" long:"help" description:"show this help message"`
		CpuProfile  string `long:"cpuprofile" description:"write pprof file" value-name:"FILE"`
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

	if opts.CpuProfile != "" {
		f, err := os.Create(opts.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if opts.NumWorkers == 0 {
		opts.NumWorkers = runtime.NumCPU()
		runtime.GOMAXPROCS(opts.NumWorkers)
	}

	if opts.NumWorkers < 0 {
		log.Fatal("value for --workers must be non-negative")
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
		projector, err := dupsquash.ParseIndices(opts.FileColumn)
		if err != nil {
			log.Fatalf("could not parse column indices: %s\n", opts.FileColumn)
		}

		queue := make(chan *Work)
		results := make(chan [][]string)
		done := make(chan bool)

		writer := bufio.NewWriter(os.Stdout)
		defer writer.Flush()
		go FanInWriter(writer, results, done)

		var wg sync.WaitGroup
		for i := 0; i < opts.NumWorkers; i++ {
			wg.Add(1)
			go Worker(queue, results, &wg)
		}

		for scanner.Scan() {
			values := strings.Split(scanner.Text(), opts.FileDelimiter)
			likeText, err := dupsquash.ConcatenateValuesNull(values, projector, opts.FileNullValue)
			if err != nil {
				log.Fatal(err)
			}

			work := Work{
				Indices:       indices,
				Connection:    conn,
				Fields:        fields,
				LikeText:      likeText,
				MinTermFreq:   opts.MinTermFreq,
				MaxQueryTerms: opts.MaxQueryTerms,
				Size:          opts.Size,
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
