package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindGoFiles(t *testing.T) {
	// Test without including test files
	files, err := findGoFiles("testdata", false, false, []string{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"testdata/simple/main.go", "testdata/comments/code.go"}, files)

	// Test with including test files
	files, err = findGoFiles("testdata", true, false, []string{})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"testdata/simple/main.go", "testdata/simple/main_test.go", "testdata/comments/code.go"}, files)
}

func TestCountLines(t *testing.T) {
	content := `package main

// A comment
import "fmt"

/*
Another
comment
*/
func main() {
	fmt.Println("Hello") // Line with comment
}
`
	lines, loc := countLines(content)
	assert.Equal(t, 12, lines)
	assert.Equal(t, 5, loc)
}

func TestAnalyzeFile(t *testing.T) {
	stats, err := analyzeFile("testdata/simple/main.go")
	assert.NoError(t, err)
	assert.Equal(t, "testdata/simple/main.go", stats.Path)
	assert.Equal(t, 16, stats.Lines)
	assert.Equal(t, 12, stats.LOC)
	assert.Equal(t, 2, stats.Functions)
	assert.Equal(t, 1, stats.Structs)

	stats, err = analyzeFile("testdata/comments/code.go")
	assert.NoError(t, err)
	assert.Equal(t, 16, stats.Lines)
	assert.Equal(t, 6, stats.LOC)
	assert.Equal(t, 1, stats.Functions)
	assert.Equal(t, 1, stats.Structs)
}
