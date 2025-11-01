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

func isMockOrGenerated(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, "_mock.go") ||
		strings.HasSuffix(base, "_gen.go") ||
		strings.HasSuffix(base, "_generated.go") ||
		strings.HasPrefix(base, "mock_") ||
		strings.HasPrefix(base, "generated_") {
		return true
	}
	// ディレクトリ名判定
	dirs := strings.Split(path, string(os.PathSeparator))
	for _, d := range dirs {
		if strings.Contains(strings.ToLower(d), "mock") || strings.Contains(strings.ToLower(d), "generated") || strings.Contains(strings.ToLower(d), "gen") {
			return true
		}
	}
	return false
}

func findGoFiles(root string, includeTests bool, includeMocks bool, excludeDirs []string) ([]string, error) {
	var files []string
	excludeMap := make(map[string]struct{})
	for _, d := range excludeDirs {
		if d != "" {
			excludeMap[d] = struct{}{}
		}
	}
	rootAbs, _ := filepath.Abs(root)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			rel, _ := filepath.Rel(rootAbs, path)
			parts := strings.Split(rel, string(os.PathSeparator))
			for _, p := range parts {
				if _, ok := excludeMap[p]; ok {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if !includeMocks && isMockOrGenerated(path) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func printStats(packages map[string]*PackageStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
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

		fmt.Fprintf(w, "%-s\t%d\t%d\t%d\t%d\t%d\t%.1f\t\n",
			pkg.Name, pkg.Files, pkg.Lines, pkg.LOC, pkg.Functions, pkg.Structs, avgLoc)
	}

	// Separator
	fmt.Fprintln(w, " \t \t \t \t \t \t \t")

	// Total
	totalAvgLoc := 0.0
	if total.Files > 0 {
		totalAvgLoc = float64(total.LOC) / float64(total.Files)
	}
	fmt.Fprintf(w, "%-s\t%d\t%d\t%d\t%d\t%d\t%.1f\t\n",
		total.Name, total.Files, total.Lines, total.LOC, total.Functions, total.Structs, totalAvgLoc)
}

func main() {
	// Define flags
	includeTests := flag.Bool("tests", false, "include test files in statistics")
	includeMocks := flag.Bool("mocks", false, "include mock/generated files in statistics")
	excludeDirs := flag.String("exclude", "vendor", "comma-separated list of directories to exclude")
	depth := flag.Int("depth", -1, "limit directory depth for stats grouping (-1 for unlimited)")
	flag.Parse()

	// Determine the root directory from positional arguments
	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	excludeList := strings.Split(*excludeDirs, ",")
	for i := range excludeList {
		excludeList[i] = strings.TrimSpace(excludeList[i])
	}

	goFiles, err := findGoFiles(root, *includeTests, *includeMocks, excludeList)
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

	packages := aggregateStats(allStats, root, *depth)
	printStats(packages)
}
