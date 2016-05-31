package cnfhash

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"strconv"
)

const literalDelim = " "
const clauseDelim = "0\n"

func HashCNF(in <-chan int64, out chan<- string) {
	var hasher hash.Hash = sha1.New()
	i := 0
	for integer := range in {
		if i < 2 {
			i++
			continue
		}
		strRepr := strconv.FormatInt(integer, 10)

		if integer == 0 {
			io.WriteString(hasher, clauseDelim)
		} else {
			io.WriteString(hasher, strRepr)
			io.WriteString(hasher, literalDelim)
		}
	}
	out <- "cnf2$" + hex.EncodeToString(hasher.Sum(nil))
}

func HashDIMACS(in io.Reader, conf Config) (string, error) {
	intChan := make(chan int64)
	errChan := make(chan error)
	outChan := make(chan string)

	go ParseDimacsFileIntegers(in, intChan, errChan, conf)
	go HashCNF(intChan, outChan)

	for {
		select {
		case e := <-errChan:
			return "", e
		case o := <-outChan:
			close(errChan)
			return o, nil
		}
	}
}
