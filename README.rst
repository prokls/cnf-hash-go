cnf-hash-go
===========

:author:        Lukas Prokop
:date:          August 2015, May 2016
:version:       1.1.0
:license:       CC-0

A Go implementation to hash CNF/DIMACS files.
See `the technical report <http://lukas-prokop.at/proj/megosat/downloads/cnf-hash.pdf>`_ for more details.

How to use
----------

Use ``go get`` to fetch the package::

    go get github.com/prokls/cnf-hash-go

Then use this import URL to use the package. Example::

    package main

    import (
        "fmt"
        "os"

        "github.com/prokls/cnf-hash-go/cnfhash"
    )

    func main() {
        // Parse DIMACS files and hash them
        fd, err := os.Open("test.cnf")
        if err != nil {
            panic(err)
        }

        fmt.Println(cnfhash.HashDIMACS(fd, []string{"c"}))

        // or hash integers directly

        inChan := make(chan int64)
        outChan := make(chan string)
        ints := []int64{3, 2, 1, -3, 0, -1, 2, 0}

        go cnfhash.HashCNF(inChan, outChan)
        for _, i := range ints {
            inChan <- i
        }
        close(inChan)

        hashValue := <-outChan
        fmt.Println(hashValue)
    }

Testing the software
--------------------

Download `the testsuite <http://github.com/prokls/cnf-hash-tests1/>`_.
Provide the folder location as environment variable::

    export TESTSUITE="/home/prokls/Downloads/cnf-hash/tests1/"

Then test the main test file::

    go test main_test.go

The testsuite has been run successfully, if the exit code is 0.
This implementation cannot handle literals exceeding 2^64 signed integer
and therefore corresponding testcases are skipped.

DIMACS file assumptions
-----------------------

A DIMACS file is valid iff

1. Any line starting with "c" or consisting only of whitespace is considered as *comment line* and content is not interpreted until the next newline character occurs.
2. The remaining file is a sequence of whitespace separated values.

   1. The first value is required to be "p"
   2. The second value is required to be "cnf"
   3. The third value is required to be a non-negative integer and called *nbvars*.
   4. The fourth value is required to be a non-negative integer and called *nbclauses*.
   5. The remaining non-zero integers are called *lits*.
   6. The remaining zero integers are called *clause terminators*.

3. A DIMACS file must be terminated by a clause terminator.
4. Every lit must satisfy ``-nbvars ≤ lit ≤ nbvars``.
5. The number of clause terminators must equate nbclauses.

This implementation will only accept literals which fit into a signed 64bit integer.

============== =========================================
**term**       **ASCII mapping**
-------------- -----------------------------------------
"c"            U+0063
"p"            U+0070
"cnf"          U+0063 U+006E U+0066 U+0020
sign           U+002D
nonzero digit  U+0031 – U+0039
digits         U+0030 – U+0039
whitespace     U+0020, U+000D, U+000A, U+0009
zero           U+0030
============== =========================================

Formal specification
--------------------

A valid DIMACS file is read in and a SHA1 instance is fed with bytes:

1. The first four values are dropped.
2. Lits are treated as integers without leading zeros. Integers are submitted as base 10 ASCII digits with an optional leading sign to the SHA1 instance.
3. Clause terminators are submitted as zero character followed by a newline character to the SHA1 instance.

Performance and memory
----------------------

The DIMACS parser uses OS' page size as default block size.
It uses channels to send parsed integer values to the hashing algorithm.
Besides that memory consumption is kept really low as tests also indicated.

The technical report shows that 45 DIMACS files summing up to 1~GB memory
can be read in 1225~seconds. Hence it is twice as fast as the equivalent
`Python implementation <http://github.com/prokls/cnf-hash-py/>`_
(but the Python implementation supports lit above 2^64 though).

Example
-------

::

    % cat test.cnf
    p cnf 5 6
    1 2 3 0
    2 3 -4 0
    1 -2 0
    -1 2 0
    1 3 5 0
    1 -4 -5 0
    % cnf-hash-go test.cnf
    cnf-hash 1.1.0 2016-05-29T14:01:41Z /root
    cnf1$7ca8bbcc091459201571acc083fbde4f7b1fcc94  test.cnf
    %

Cheers!
