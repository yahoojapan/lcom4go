package lcom4_test

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	lcom4 "github.com/yahoojapan/lcom4go"
)

func TestAnalyzer(t *testing.T) {
	testdata, _ := filepath.Abs("testdata")
	analysistest.Run(t, testdata, lcom4.Analyzer, "a")
}
