package main

import (
	lcom4 "github.com/yahoojapan/lcom4go"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(lcom4.Analyzer) }
