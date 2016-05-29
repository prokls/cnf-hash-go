package main

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/prokls/cnf-hash-go/cnfhash"
)

func checkFile(t *testing.T, w *sync.WaitGroup, filepath string) {
	defer w.Done()

	fd, err := os.Open(filepath)
	if err != nil {
		t.Fatal(err)
	}

	defer fd.Close()

	hashValue, err := cnfhash.HashDIMACS(fd, []string{"c"})
	if len(hashValue) > 6 && strings.HasPrefix(hashValue, "cnf1$") {
		hashValue = hashValue[5:]
	}
	// skip test cases testing value >=2^64
	if err != nil && (strings.Contains(err.Error(), "value out of range") ||
		strings.Contains(err.Error(), "supports")) {
		t.Log(err.Error())
		return
	} else if err != nil {
		t.Fatal(err)
	}

	var parts []string = strings.Split(path.Base(filepath), "_")
	var expected string = parts[0]
	if expected != hashValue {
		t.Error("Expected hash value " + expected + ", got " + hashValue)
	}
}

func TestCorrectness(t *testing.T) {
	testdir := "test"
	if os.Getenv("TESTSUITE") != "" {
		testdir = os.Getenv("TESTSUITE")
	}

	files, err := ioutil.ReadDir(testdir)
	if err != nil {
		t.Fatal(err)
	}

	var wait sync.WaitGroup
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".cnf") {
			continue
		}

		wait.Add(1)
		go checkFile(t, &wait, path.Join(testdir, file.Name()))
	}

	wait.Wait()
}
