package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Json2csv struct {
	nameSep      string
	listChar     string
	listIndex    string
	loops        map[string]int
	counters     map[string]int
	row          map[string]string
	columnNames  []string
	counterNames []string
}

func (jc *Json2csv) errorExit(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	_, _ = fmt.Fprintln(os.Stderr, jc)
	panic(err)
}

var removeDuplicateRows = true

func main() {

	var jtree interface{}
	flag.BoolVar(&removeDuplicateRows, "no-dedupe", true, "Do not remove duplicate rows.")
	flag.Parse()

	j2c := NewJson2CSV(".", "LIST", "#")

	jsonFile := os.Stdin
	defer func() { _ = jsonFile.Close() }()
	csvw := csv.NewWriter(os.Stdout)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		j2c.errorExit(err)
	}
	err = json.Unmarshal(byteValue, &jtree)
	if err != nil {
		j2c.errorExit(err)
	}

	headers := findHeaders(j2c.nameSep, j2c.listChar, j2c.listIndex, []string{}, jtree)
	j2c.setColumnNames(headers)
	err = csvw.Write(j2c.columnNames)
	if err != nil {
		j2c.errorExit(err)
	}
	j2c.findCounters([]string{}, jtree)
	j2c.resetCounters(j2c.counterNames)
	for {
		j2c.resetRow()
		skips := j2c.populateRow([]string{}, jtree)
		if skips == 0 {
			record := j2c.getRecordFromRow()
			if !rowDuplicated(record) {

				err := csvw.Write(record)
				if err != nil {
					j2c.errorExit(err)
				}
				csvw.Flush()
			}
		}
		if j2c.incrementCounters(j2c.counterNames) == true {
			return
		}
	}

}

var emittedRows = map[string]bool{}

func rowDuplicated(record []string) bool {
	if removeDuplicateRows {
		r := strings.Join(record, "\x00")
		if _, ok := emittedRows[r]; ok {
			// it's a duplicate row so leave it out
			return true
		}
		emittedRows[r] = true
	}
	return false
}

func NewJson2CSV(ns string, lc string, li string) *Json2csv {
	jc := Json2csv{
		nameSep:      ns,
		listChar:     lc,
		listIndex:    li,
		loops:        map[string]int{},
		counters:     map[string]int{},
		row:          map[string]string{},
		columnNames:  []string{},
		counterNames: []string{},
	}
	return &jc
}
func (jc *Json2csv) resetRow() {
	for k := range jc.row {
		jc.row[k] = ""
	}
}
func (jc *Json2csv) resetCounters(names []string) {
	for _, n := range names {
		jc.counters[n] = 1
	}
}

func (jc *Json2csv) incrementCounters(names []string) bool {
	if len(names) == 0 {
		return true
	}
	carry := jc.incrementCounters(names[1:])
	if carry == false {
		return false
	}
	count := jc.counters[names[0]]
	count += 1
	if count > jc.loops[names[0]] {
		return true
	}
	jc.counters[names[0]] = count
	jc.resetCounters(names[1:])
	return false
}

func findHeaders(nameSep string, listChar string, listIndex string, path []string, jtree interface{}) map[string]bool {
	pathString := strings.Join(path, nameSep)
	headers := map[string]bool{}
	switch x := jtree.(type) {
	case []interface{}:
		np := append(path, listChar)
		npath := strings.Join(np, nameSep) // e.g. users.$
		headers[npath] = true
		indexpath := strings.Join(append(np, listIndex), nameSep) // e.g. users.$.index
		headers[indexpath] = true
		for _, subtree := range x {
			subHeaders := findHeaders(nameSep, listChar, listIndex, np, subtree)
			for k, i := range subHeaders {
				headers[k] = i
			}
		}
	case map[string]interface{}:
		keyNumber := 0
		for k, subtree := range x {
			keyNumber += 1
			subHeaders := findHeaders(nameSep, listChar, listIndex, append(path, k), subtree)
			for k := range subHeaders {
				headers[k] = true
			}
		}
	default:
		headers = map[string]bool{pathString: true}
	}
	return headers
}
func (jc *Json2csv) findCounters(path []string, jtree interface{}) {
	switch x := jtree.(type) {
	case []interface{}:
		np := append(path, jc.listChar)
		ip := append(np, jc.listIndex)
		ipath := strings.Join(ip, jc.nameSep) // e.g.users.$.index
		m, ok := jc.loops[ipath]
		if !ok {
			jc.loops[ipath] = len(x)
		}
		if ok && m < len(x) {
			jc.loops[ipath] = len(x)
		}
		for _, subtree := range x {
			jc.findCounters(np, subtree)
		}
	case map[string]interface{}:
		for k, subtree := range x {
			jc.findCounters(append(path, k), subtree)
		}
	default:

	}
	if len(path) == 0 {
		jc.counterNames = []string{}
		for k := range jc.loops {
			jc.counterNames = append(jc.counterNames, k)
		}
		sort.Strings(jc.counterNames)
	}
}

func (jc *Json2csv) populateRow(path []string, jtree interface{}) int {
	skipCount := 0
	pathString := strings.Join(path, jc.nameSep)
	switch x := jtree.(type) {
	case []interface{}:
		np := append(path, jc.listChar)
		ip := append(np, jc.listIndex)
		ipath := strings.Join(ip, jc.nameSep)
		index := jc.counters[ipath]
		jc.row[ipath] = strconv.Itoa(index)
		if len(x) < index {
			jc.populateRow(np, "")
			return skipCount + 1
		}
		skipCount += jc.populateRow(np, x[index-1])
	case map[string]interface{}:
		for k, subtree := range x {
			skipCount += jc.populateRow(append(path, k), subtree)
		}
	default:
		jc.row[pathString] = fmt.Sprintf("%v", jtree)
	}
	return skipCount
}

func (jc *Json2csv) setColumnNames(headers map[string]bool) {
	jc.columnNames = []string{}
	for k := range headers {
		jc.columnNames = append(jc.columnNames, k)
	}
	valueFn := func(i int) int {
		n := jc.columnNames[i]
		return 100*(strings.Count(n, jc.nameSep)-strings.Count(n, jc.listIndex)) + len(n)
	}
	sort.Slice(jc.columnNames, func(i, j int) bool {
		return valueFn(i) < valueFn(j)
	})
	values := map[string]int{}
	for i, v := range jc.columnNames {
		values[v] = valueFn(i)
	}

}

func (jc *Json2csv) getRecordFromRow() []string {
	var record []string
	for _, colName := range jc.columnNames {
		record = append(record, jc.row[colName])
	}
	return record
}
