//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <coverage-file> <threshold>\n", os.Args[0])
		os.Exit(1)
	}

	coverageFile := os.Args[1]
	threshold, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid threshold: %v\n", err)
		os.Exit(1)
	}

	coverage, err := parseCoverage(coverageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total coverage: %.1f%%\n", coverage)
	if coverage < threshold {
		fmt.Fprintf(os.Stderr, "Coverage %.1f%% is below %.0f%% threshold\n", coverage, threshold)
		os.Exit(1)
	}
	fmt.Printf("Coverage %.1f%% meets %.0f%% threshold\n", coverage, threshold)
}

func parseCoverage(filename string) (float64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				pctStr := strings.TrimSuffix(fields[2], "%")
				return strconv.ParseFloat(pctStr, 64)
			}
		}
	}
	return 0, fmt.Errorf("no coverage total found in %s", filename)
}
