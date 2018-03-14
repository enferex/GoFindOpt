package main

import (
	"bytes"
	"debug/elf"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

type ElfObj struct {
	getOptStrings []string
	fileName      string
	isValid       bool
	minOptstrLen  int // Minimum length of optstring to accept.
}

func (eo *ElfObj) dump() {
	var isValid string
	if eo.isValid {
		isValid = "Valid ELF Object"
	} else {
		isValid = "Invalid ELF Object"
	}
	for _, str := range eo.getOptStrings {
		println(eo.fileName, ",", str, ",", isValid)
	}
}

// Scan the list looking for a string that starts with 'getopt'
func searchForGetOptSymbol(ch chan bool, symbols []elf.Symbol) {
	found := false
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, "getopt") {
			found = true
			break
		}
	}
	ch <- found
}

// Scan all symbols looking for getopt.  If that
// symbol is found, then return true..
func hasGetOptSymbol(file *elf.File) bool {
	const nSymLists = 2
	ch := make(chan bool, nSymLists)
	defer close(ch)
	found := false
	syms, _ := file.Symbols()        // List 1
	dyms, _ := file.DynamicSymbols() // List 2
	go searchForGetOptSymbol(ch, syms)
	go searchForGetOptSymbol(ch, dyms)
	for count := 0; count < nSymLists; count++ {
		if ok := <-ch; ok == true {
			// Found the symbol, so exit early.
			found = true
		}
	}
	return found
}

// Return the next string starting from idx.
func nextString(startIdx int, data []byte) (int, string) {
	println("[", startIdx, ",", len(data), "]")
	for idx := startIdx; idx < len(data); idx++ {
		if data[idx] == 0 {
			return idx, string(data[startIdx:idx])
		}
	}
	return -1, ""
}

// Check all alpha characters and ensure there are no duplicates.
func hasDuplicateChars(str string) bool {
	exists := make(map[rune]bool, len(str))
	for _, c := range str {
		if exists[c] == true {
			return true
		}
		exists[c] = true
	}
	return false
}

// Look for getopt like strings in .rodata.
func findStringTableMatch(file *elf.File, re *regexp.Regexp) []string {
	possibleGetOptStrings := []string{}
	rodata := file.Section(".rodata")
	if rodata == nil {
		return []string{"N/A"}
	}

	// We have a .rodata, read in the section contents.
	data, err := rodata.Data()
	if err != nil {
		return []string{"N/A"}
	}

	// Scan all the things looking for string like things.
	for _, str := range bytes.Split(data, []byte("\x00")) {
		if re.MatchString(string(str)) && !hasDuplicateChars(string(str)) {
			possibleGetOptStrings = append(possibleGetOptStrings, string(str))
		}
	}

	if len(possibleGetOptStrings) == 0 {
		return []string{"N/A"}
	}

	return possibleGetOptStrings
}

// Convert the filename into an ElfObj, this includes
// scanning for getopt strings.
func NewElfObj(elfObj *ElfObj, filename string, min int) {
	*elfObj = ElfObj{fileName: filename, isValid: false, minOptstrLen: min}
	file, err := elf.Open(filename)
	if err != nil {
		elfObj.getOptStrings = []string{"N/A"}
		elfObj.isValid = false
		return
	}

	elfObj.isValid = true

	if hasGetOptSymbol(file) {
		// See: man 3 getopt  -- In particular the 'optstring' definition.
		// First character can be an optional '+' or '-'.
		// At least one A-Z a-z with optional ':' between them.
		// This is a subset of what getopt permits.
		restr := fmt.Sprintf("^[+-]?([a-zA-Z]+:?){%v,}$", elfObj.minOptstrLen)
		re := regexp.MustCompile(restr)
		elfObj.getOptStrings = findStringTableMatch(file, re)
	} else {
		elfObj.getOptStrings = []string{"N/A"}
	}
}

func main() {
	min := flag.Int("n", 2, "Minimum suspected optstring length to search for.")
	flag.Usage = func() {
		println("Usage: ", os.Args[0], "<-n num> [elf...]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	// Treat the arguments passed via command line as a filename to be scanned.
	elfs := make([]ElfObj, len(flag.Args()), len(flag.Args()))
	var wg sync.WaitGroup
	for i, arg := range flag.Args() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			NewElfObj(&elfs[i], arg, *min)
		}()
	}
	wg.Wait()

	// Done scanning all files.
	for _, elfObj := range elfs {
		elfObj.dump()
	}
}
