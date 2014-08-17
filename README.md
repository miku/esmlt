dupsquash
=========

A fuzzy deduplication tool tailored towards MARC.

dsqpool
-------

    $ dsqpool -h
    Usage:
      dsqpool [OPTIONS]

    Application Options:
          --host=HOST               elasticsearch host (localhost)
          --port=PORT               elasticsearch port (9200)
      -l, --like=STRING             string to compare
          --file=FILENAME           onput file (TSV) with strings to compare
      -f, --column=COLUMN[S]        which column(s) to pick for the comparison (1)
          --index=NAME[S]           index or indices (space separated)
      -x, --index-fields=NAME[S]    which index fields to use for comparison
                                    (content.245.a content.245.b)
          --min-term-freq=N         passed on lucene option (1)
          --max-query-terms=N       passed on lucene option (25)
      -s, --size=N                  number of results per query (5)
      -V, --version                 show version and exit (false)
      -h, --help                    show this help message (false)

Start simple with the `--like` option. This will run a *more like this* query, which will
return document ids, that are similar (in title and subtitle) to the word "Language":

    $ go run cmd/dsqpool/dsqpool.go --like Language
    ebl EBL1599295      6.759
    ebl EBL1715045      6.382
    nep 9781108063784   5.320
    nep 9780415576826   5.270
    nep 9780415487528   5.270
