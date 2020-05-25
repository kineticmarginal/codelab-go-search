package main

import (
"bufio"
"flag"
"fmt"
"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	recursiveFlag = flag.Bool("r", false, "recursive search: for directories")
	lineNumber = flag.Bool("n", false, "print line number")
)

type ScanResult struct {
	file       string
	lineNumber int
	line       string
}

func exit(format string, val ...interface{}) {
	if len(val) == 0 {
		fmt.Println(format)
	} else {
		fmt.Printf(format, val)
		fmt.Println()
	}
	os.Exit(1)
}

func scanFile(fpath, pattern string) ([]ScanResult) {
	f, err := os.Open(fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	result := make([]ScanResult, 0)
	lineN := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			result = append(result, ScanResult{
				file:       fpath,
				lineNumber: lineN,
				line:       line,
			})
		}
		lineN++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return result
}

func processFile(fpath string, pattern string) {
	res := scanFile(fpath, pattern)
	for _, arg := range res {
		if *lineNumber {
			fmt.Printf("%s:%d:%s\n", arg.file, arg.lineNumber, arg.line)
		} else {
			fmt.Printf("%s:%s\n", arg.file, arg.line)
		}
	}
}

func processDirectory(dir string, pattern string) chan []ScanResult{
	res := make(chan []ScanResult)
	go func() {
		var wg sync.WaitGroup
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			wg.Add(1)
			go func() {
				res <- scanFile(path, pattern)
				wg.Done()
			}()
			return nil
		})
		go func() {
			wg.Wait()
			close(res)
		}()
	} ()
	return res
}

func printDir(res chan []ScanResult) {
	for arg := range res {
		for i := range arg {
			if *lineNumber {
				fmt.Printf("%s:%d:%s\n", arg[i].file, arg[i].lineNumber, arg[i].line)
			} else {
				fmt.Printf("%s:%s\n", arg[i].file, arg[i].line)
			}
		}
	}
}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		exit("usage: go-search <path> <pattern> to search")
	}

	path := flag.Arg(0)
	pattern := flag.Arg(1)

	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	recursive := *recursiveFlag
	if info.IsDir() && !recursive {
		exit("%s: is a directory", info.Name())
	}

	start := time.Now()
	if info.IsDir() && recursive {
		result := processDirectory(path, pattern)
		printDir(result)
	} else {
		processFile(path, pattern)
	}
	fmt.Println("Elapsed: ", time.Now().Sub(start))
}
