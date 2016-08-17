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
		fmt.Fprintf(os.Stderr, "error opening file '%s'\n", dimacsFile)
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		return ""
	}
	defer fd.Close()
	hashvalue, err := cnfhash.HashDIMACS(fd, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error computing hash value for '%s'\n", dimacsFile)
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		return ""
	}
	return hashvalue
}

func runGzip(archive string, conf cnfhash.Config) string {
	gzipFd, err := os.Open(archive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening '%s'\n", archive)
		fmt.Fprintf(os.Stderr, err.Error())
		return ""
	}
	defer gzipFd.Close()

	fd, err := gzip.NewReader(gzipFd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening '%s'\n", archive)
		fmt.Fprintf(os.Stderr, err.Error())
		return ""
	}
	defer fd.Close()

	hashvalue, err := cnfhash.HashDIMACS(fd, conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error computing hash value in archive '%s'\n", archive)
		fmt.Fprintf(os.Stderr, err.Error())
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
	checkHeader := false
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if ignore {
			ignoreLines = append(ignoreLines, arg)
			ignore = false
		} else if arg == "--fullpath" || arg == "-f" {
			fullPath = true
		} else if arg == "--ignore" {
			ignore = true
		} else if arg == "--header-check" {
			checkHeader = true
		} else {
			dimacsFiles = append(dimacsFiles, arg)
		}
	}

	// usage information
	if len(dimacsFiles) == 0 {
		fmt.Println("cnf-hash-go [--fullpath] [--ignore C] [--header-check] [file.cnf]+")
		fmt.Println()
		fmt.Println("  --fullpath      print full path of CNF file, not basename")
		fmt.Println("  --header-check  check values of CNF header")
		fmt.Println("  --ignore C      ignore all lines starting with character C")
		fmt.Println("  file.cnf        DIMACS files representing some CNF")
		fmt.Println()

		os.Exit(1)
	}

	var wait sync.WaitGroup
	success := true
	timestamp := time.Now().UTC().Format(time.RFC3339)
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error determining working directory")
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	fmt.Println("cnf-hash 2.1.1 " + timestamp + " " + pwd)
	for i := 0; i < len(dimacsFiles); i++ {
		wait.Add(1)
		go func(dimacsFile string, ignoreLines []string) {
			var hashValue string
			var conf cnfhash.Config
			conf.CheckHeader = checkHeader
			conf.IgnoreLines = ignoreLines
			if strings.HasSuffix(dimacsFile, ".gz") {
				hashValue = runGzip(dimacsFile, conf)
			} else {
				hashValue = runFile(dimacsFile, conf)
			}
			if hashValue == "" {
				success = false
				fmt.Fprintf(os.Stderr, "error while processing file '%s' - skipping file\n", dimacsFile)
			} else {
				if fullPath {
					fmt.Printf("%s  %s\n", hashValue, dimacsFile)
				} else {
					fmt.Printf("%s  %s\n", hashValue, path.Base(dimacsFile))
				}
			}
			wait.Done()
		}(dimacsFiles[i], ignoreLines)
	}
	wait.Wait()

	if !success {
		os.Exit(2)
	}
}
