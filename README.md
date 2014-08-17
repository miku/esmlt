dupsquash
=========

A fuzzy deduplication tool tailored towards MARC.

[![Build Status](http://img.shields.io/travis/miku/dupsquash.svg?style=flat)](https://travis-ci.org/miku/dupsquash)

dsqpool
-------

    $ duppool -h
    Usage:
      duppool [OPTIONS]

    Application Options:
          --host=HOST               elasticsearch host (localhost)
          --port=PORT               elasticsearch port (9200)
      -l, --like=STRING             string to compare
      -i, --file=FILENAME           input file (TSV) with strings to compare
      -f, --column=COLUMN[S]        which column(s) to pick for the comparison (1)
          --delimiter=DELIM         column delimiter of the file (  )
          --null-value=STRING       value that indicates empty value in input file (<NULL>)
          --index=NAME[S]           index or indices (space separated)
      -x, --index-fields=NAME[S]    which index fields to use for comparison
                                    (content.245.a content.245.b)
          --min-term-freq=N         passed on lucene option (1)
          --max-query-terms=N       passed on lucene option (25)
      -s, --size=N                  number of results per query (5)
      -w, --workers=N               number of workers, 0 means number of available cpus (0)
      -V, --version                 show version and exit (false)
      -h, --help                    show this help message (false)
          --cpuprofile=FILE         write pprof file

Start simple with the `--like` option. This will run a *more like this* query
over all indices. It will return document ids, that are similar
(in title and subtitle) to the word "Language":

    $ duppool --like Language
    ebl title   EBL1599295  6.759
    ebl title   EBL1715045  6.382
    nep title   9781108063784   5.320
    nep title   9780415576826   5.270
    nep title   9780415487528   5.270
