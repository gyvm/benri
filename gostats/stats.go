package main

import (
	"bufio"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// FileStats holds the statistics for a single Go source file.
type FileStats struct {
	Path      string
	Lines     int
	LOC       int
	Functions int
	Structs   int
}

// PackageStats holds aggregated statistics for a package (directory).
type PackageStats struct {
	Name      string
	Files     int
	Lines     int
	LOC       int
	Functions int
	Structs   int
}

// analyzeFile parses a single Go file and returns its statistics.
func analyzeFile(path string) (*FileStats, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines, loc := countLines(string(content))

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	stats := &FileStats{
		Path:  path,
		Lines: lines,
		LOC:   loc,
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncDecl:
			stats.Functions++
		case *ast.TypeSpec:
			if _, ok := n.(*ast.TypeSpec).Type.(*ast.StructType); ok {
				stats.Structs++
			}
		}
		return true
	})

	return stats, nil
}

// countLines counts total lines and lines of code (LOC).
func countLines(content string) (totalLines, loc int) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	inBlockComment := false
	for scanner.Scan() {
		totalLines++
		line := strings.TrimSpace(scanner.Text())

		if inBlockComment {
			if strings.Contains(line, "*/") {
				inBlockComment = false
				parts := strings.SplitN(line, "*/", 2)
				if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
					loc++
				}
			}
			continue
		}

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "/*") {
			inBlockComment = true
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}

		loc++
	}
	return totalLines, loc
}

// aggregateStats groups file stats by package (directory).
func aggregateStats(files []*FileStats, root string, depth int) map[string]*PackageStats {
	packages := make(map[string]*PackageStats)
	rootAbs, _ := filepath.Abs(root)

	for _, file := range files {
		absPath, _ := filepath.Abs(file.Path)
		relPath, err := filepath.Rel(rootAbs, absPath)
		if err != nil {
			relPath = file.Path // fallback
		}
		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = "main"
		}
		// depth制限
		if depth > 0 {
			parts := strings.Split(dir, string(os.PathSeparator))
			if len(parts) > depth {
				dir = strings.Join(parts[:depth], string(os.PathSeparator))
			}
		}
		if _, ok := packages[dir]; !ok {
			packages[dir] = &PackageStats{Name: dir}
		}
		pkg := packages[dir]
		pkg.Files++
		pkg.Lines += file.Lines
		pkg.LOC += file.LOC
		pkg.Functions += file.Functions
		pkg.Structs += file.Structs
	}
	return packages
}
