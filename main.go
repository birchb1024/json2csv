package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

var loops = map[string]int{}
var counters = map[string]int{}
var row = map[string]string{}

var nameSep string = "."
var listChar string = "$"

func main() {
	var jsonTree interface{}
	var jtree interface{}

	jsonFile := os.Stdin
	defer jsonFile.Close()
	csvw := csv.NewWriter(os.Stdout)

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = json.Unmarshal(byteValue, &jsonTree)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	jtree = jsonTree
	headers := findHeaders([]string{}, jtree)

	columnNames := []string{}
	for k := range headers {
		columnNames = append(columnNames, k)
	}
	sort.Strings(columnNames)
	csvw.Write(columnNames)
	findCounters([]string{}, jtree)
	counterNames := []string{}
	for k := range loops {
		counterNames = append(counterNames, k)
	}
	sort.Strings(counterNames)
	resetCounters(counterNames)
	for {
		//for _, n := range counterNames {
		//	fmt.Print(n, "=", counters[n], " ")
		//}
		//fmt.Println()
		resetRow()
		populateRow([]string{}, jtree)
		var record []string
		for _, colName := range columnNames {
			record = append(record, row[colName])
		}
		err := csvw.Write(record)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		csvw.Flush()
		if incrementCounters(counterNames) == true {
			return
		}
	}

}

func resetRow() {
	for k := range row {
		row[k] = ""
	}
}
func resetCounters(names []string) {
	for _, n := range names {
		counters[n] = 1
	}
}

func incrementCounters(names []string) bool {
	if len(names) == 0 {
		return true
	}
	carry := incrementCounters(names[1:])
	if carry == false {
		return false
	}
	count := counters[names[0]]
	count += 1
	if count > loops[names[0]] {
		return true
	}
	counters[names[0]] = count
	resetCounters(names[1:])
	return false
}

func findHeaders(path []string, jtree interface{}) map[string]bool {
	pathString := strings.Join(path, nameSep)
	headers := map[string]bool{}
	switch x := jtree.(type) {
	case []interface{}:
		np := append(path, listChar)
		npath := strings.Join(np, nameSep)
		headers[npath] = true
		for _, subtree := range x {
			subHeaders := findHeaders(np, subtree)
			for k, i := range subHeaders {
				headers[k] = i
			}
		}
	case map[string]interface{}:
		keyNumber := 0
		for k, subtree := range x {
			keyNumber += 1
			subHeaders := findHeaders(append(path, k), subtree)
			for k := range subHeaders {
				headers[k] = true
			}
		}
	default:
		headers = map[string]bool{pathString: true}
	}
	return headers
}
func findCounters(path []string, jtree interface{}) {
	switch x := jtree.(type) {
	case []interface{}:
		np := append(path, listChar)
		npath := strings.Join(np, nameSep)
		m, ok := loops[npath]
		if !ok {
			loops[npath] = len(x)
		}
		if ok && m < len(x) {
			loops[npath] = len(x)
		}
		for _, subtree := range x {
			findCounters(np, subtree)
		}
	case map[string]interface{}:
		for k, subtree := range x {
			findCounters(append(path, k), subtree)
		}
	default:

	}
}

func populateRow(path []string, jtree interface{}) {
	pathString := strings.Join(path, nameSep)
	switch x := jtree.(type) {
	case []interface{}:
		np := append(path, listChar)
		npath := strings.Join(np, nameSep)
		index := counters[npath]
		row[npath] = strconv.Itoa(index)
		if len(x) >= index {
			populateRow(np, x[index-1])
		}
	case map[string]interface{}:
		for k, subtree := range x {
			populateRow(append(path, k), subtree)
		}
	default:
		row[pathString] = fmt.Sprintf("%v", jtree)
	}
}
