package main

import (
	"compress/gzip"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/prokls/cnf-hash-go/cnfhash"
)

func runFile(dimacsFile string, conf cnfhash.Config) string {
	fd, err := os.Open(dimacsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening "+dimacsFile, err)
		return ""
	}
	defer fd.Close()
	hashvalue, err := cnfhash.HashDIMACS(fd, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error computing hash value for "+dimacsFile, err)
		return ""
	}
	return hashvalue
}

func runGzip(archive string, conf cnfhash.Config) string {
	gzipFd, err := os.Open(archive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening "+archive, err)
		return ""
	}
	defer gzipFd.Close()

	fd, err := gzip.NewReader(gzipFd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening "+archive, err)
		return ""
	}
	defer fd.Close()

	hashvalue, err := cnfhash.HashDIMACS(fd, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error computing hash value in archive "+archive, err)
		return ""
	}

	return hashvalue
}

func main() {
	fullPath := false
	ignoreLines := make([]string, 2)
	ignoreLines = append(ignoreLines, "c")
	ignoreLines = append(ignoreLines, "%")

	dimacsFiles := make([]string, 0, 5)

	// manual CLI parsing, because I dislike the "flag" package
	ignore := false
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if ignore {
			ignoreLines = append(ignoreLines, arg)
			ignore = false
		} else if arg == "--fullpath" || arg == "f" {
			fullPath = true
		} else if arg == "--ignore" {
			ignore = true
		} else {
			dimacsFiles = append(dimacsFiles, arg)
		}
	}

	// usage information
	if len(dimacsFiles) == 0 {
		fmt.Println("cnf-hash-go [--fullpath] [--ignore C] [file.cnf]+")
		fmt.Println()
		fmt.Println("  --fullpath      print full path of CNF file, not basename")
		fmt.Println("  --ignore C      ignore all lines beginning with C")
		fmt.Println("  file.cnf        DIMACS files representing some CNF")
		fmt.Println()

		os.Exit(1)
	}

	var wait sync.WaitGroup
	success := true
	timestamp := time.Now().UTC().Format(time.RFC3339)
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error determining working directory", err)
		return
	}

	fmt.Println("cnf-hash 1.0.0 " + timestamp + " " + pwd)
	for i := 0; i < len(dimacsFiles); i++ {
		wait.Add(1)
		go func(dimacsFile string, ignoreLines []string) {
			var hashValue string
			var conf cnfhash.Config
			conf.IgnoreLines = ignoreLines
			if strings.HasSuffix(dimacsFile, ".gz") {
				hashValue = runGzip(dimacsFile, conf)
			} else {
				hashValue = runFile(dimacsFile, conf)
			}
			if hashValue == "" {
				success = false
			}
			if fullPath {
				fmt.Printf("%s  %s\n", hashValue, dimacsFile)
			} else {
				fmt.Printf("%s  %s\n", hashValue, path.Base(dimacsFile))
			}
			wait.Done()
		}(dimacsFiles[i], ignoreLines)
	}
	wait.Wait()

	if !success {
		os.Exit(2)
	}
}
