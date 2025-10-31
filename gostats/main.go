package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

func findGoFiles(root string, includeTests bool) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func printStats(packages map[string]*PackageStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "Name\tFiles\tLines\tLOC\tFuncs\tStructs\tAvg LOC/File\t")

	// Sort packages by name for consistent output
	var keys []string
	for k := range packages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	total := &PackageStats{Name: "Total"}

	for _, k := range keys {
		pkg := packages[k]
		total.Files += pkg.Files
		total.Lines += pkg.Lines
		total.LOC += pkg.LOC
		total.Functions += pkg.Functions
		total.Structs += pkg.Structs

		avgLoc := 0.0
		if pkg.Files > 0 {
			avgLoc = float64(pkg.LOC) / float64(pkg.Files)
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%.1f\t\n",
			pkg.Name, pkg.Files, pkg.Lines, pkg.LOC, pkg.Functions, pkg.Structs, avgLoc)
	}

	// Separator
	fmt.Fprintln(w, " \t \t \t \t \t \t \t")

	// Total
	totalAvgLoc := 0.0
	if total.Files > 0 {
		totalAvgLoc = float64(total.LOC) / float64(total.Files)
	}
	fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%d\t%.1f\t\n",
		total.Name, total.Files, total.Lines, total.LOC, total.Functions, total.Structs, totalAvgLoc)
}

func main() {
	// Define flags
	includeTests := flag.Bool("tests", false, "include test files in statistics")
	flag.Parse()

	// Determine the root directory from positional arguments
	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	goFiles, err := findGoFiles(root, *includeTests)
	if err != nil {
		log.Fatalf("error finding Go files: %v\n", err)
	}

	var allStats []*FileStats
	for _, file := range goFiles {
		stats, err := analyzeFile(file)
		if err != nil {
			log.Printf("error analyzing file %s: %v\n", file, err)
			continue
		}
		allStats = append(allStats, stats)
	}

	packages := aggregateStats(allStats)
	printStats(packages)
}
