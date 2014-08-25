esmlt
=====

Run many more-like-this queries agains elasticsearch in parallel.

[![Build Status](http://img.shields.io/travis/miku/esmlt.svg?style=flat)](https://travis-ci.org/miku/esmlt)

Installation
------------

    $ go get github.com/miku/esmlt/cmd/esmlt

Usage
-----

    $ esmlt
    Usage: esmlt [OPTIONS]
      -columns="1": which column to use as like-text
      -cpuprofile="": write cpu profile to file
      -delimiter="\t": column delimiter of the input file
      -fields="content.245.a content.245.b": index fields to query
      -file="": input file
      -host="localhost": elasticsearch host
      -indices="": index or indices to query
      -like="": more like this queries like-text
      -max-query-terms=25: max query terms
      -min-term-freq=1: min term frequency
      -null="NOT_AVAILABLE": column value to ignore
      -port="9200": elasticsearch port
      -size=5: maximum number of similar records to report
      -v=false: prints current program version
      -workers=4: number of workers to use

Start simple with the `-like` option. This will run a *more like this* query
over all indices. It will return document ids, that are similar
(in title and subtitle) to the word "Language":

    $ esmlt -like Language
