package cnfhash

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var matchInteger *regexp.Regexp
var matchComment *regexp.Regexp
var matchHeader *regexp.Regexp
var matchEmptyLine *regexp.Regexp
var matchTerminatingLine *regexp.Regexp

// initialize all regular expressions
func init() {
	var err error

	matchInteger, err = regexp.Compile("-?[0-9]+")
	if err != nil {
		panic(err)
	}

	matchHeader, err = regexp.Compile("^p\\s+cnf\\s+(\\d+)\\s+(\\d+)\\s*$")
	if err != nil {
		panic(err)
	}

	matchEmptyLine, err = regexp.Compile("^\\s*$")
	if err != nil {
		panic(err)
	}

	matchTerminatingLine, err = regexp.Compile("^\\s*%\\s*$")
	if err != nil {
		panic(err)
	}
}

// ParseDimacsFileIntegers yields integers defined in a DIMACS file.
// The first two integers represent nbvars and nbclauses. The following
// integers represent literals in clauses terminated by a zero.
// The integers are passed to the integer channel provided as argument.
// conf allows to parameterize
func ParseDimacsFileIntegers(in io.Reader, out chan<- int64, errChan chan<- error, conf Config) {
	var err error

	if conf.IgnoreLines == nil {
		conf.IgnoreLines = make([]string, 2)
		conf.IgnoreLines = append(conf.IgnoreLines, "c")
		conf.IgnoreLines = append(conf.IgnoreLines, "%")
	}

	// preparation & initialization
	matchComment, err = regexp.Compile("^(" + strings.Join(conf.IgnoreLines, "|") + ")(\\s+|$)")
	if err != nil {
		panic(err)
	}

	var nbClauses, nbVars int64
	var tmp int64
	lineno := 1
	state := 0
	clauses := int64(0)
	wasZero := false

	// read line by line
	var scanner *bufio.Scanner = bufio.NewScanner(in)

	for scanner.Scan() {
		var line string = scanner.Text()
		var errSuffix string = fmt.Sprintf(" at line %d\n", lineno)
		lineno += 1

		if matchComment.MatchString(line) || matchEmptyLine.MatchString(line) {
			continue
		} else if matchTerminatingLine.MatchString(line) {
			break
		} else if state == 0 {
			var matches []string = matchHeader.FindStringSubmatch(line)
			if len(matches) == 0 {
				errChan <- errors.New("Got invalid DIMACS header line, " +
					"expected one with 'p cnf ... ...'" + errSuffix)
				return
			}

			for i := 1; i <= 2; i++ {
				if len(matches[i]) > 20 {
					errChan <- fmt.Errorf("Implementation only supports <2^64 values")
					return
				}
			}

			nbClauses, err = strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				errChan <- err
				return
			}
			out <- nbClauses

			nbVars, err = strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				errChan <- err
				return
			}
			out <- nbVars

			state = 1
		} else if state == 1 {
			var matches [][]string = matchInteger.FindAllStringSubmatch(line, -1)
			for j := 0; j < len(matches); j++ {
				if len(matches[j][0]) > 20 {
					errChan <- fmt.Errorf("Implementation only supports <2^64 values")
					return
				}
				tmp, err = strconv.ParseInt(matches[j][0], 10, 64)
				if err != nil {
					errChan <- err
					return
				}
				if conf.CheckHeader && !(-nbVars <= tmp && tmp < nbVars) {
					errChan <- fmt.Errorf("%d outside of range %d to %d", tmp, -nbVars, nbVars)
					return
				}
				if wasZero && tmp == 0 { // ignore consecutive zeros
					continue
				}
				if tmp == 0 {
					clauses++
					wasZero = true
				} else {
					wasZero = false
				}
				out <- tmp
			}
		}
	}
	if err = scanner.Err(); err != nil {
		errChan <- err
		return
	}
	if state != 1 {
		errChan <- errors.New("Empty DIMACS file, expected at least a header")
		return
	}
	if conf.CheckHeader && clauses != nbClauses {
		errChan <- fmt.Errorf("Expected %d clauses, got %d clauses", nbClauses, clauses)
		return
	}
	if conf.CheckHeader && !wasZero {
		errChan <- errors.New("CNFs must be terminated by a zero for the last clause")
		return
	}

	close(out)
}
