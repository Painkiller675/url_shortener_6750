package main

import (
	"github.com/Painkiller675/url_shortener_6750/internal/lib/testLint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	// определяем map подключаемых правил
	checks := map[string]bool{
		"SA5000": true,
		"SA6000": true,
		"SA9004": true,
		"SA4006": true,
		"ST1000": true,
		"QF1002": true,
		"S1000":  true,
	}
	var mychecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	multichecker.Main(
		mychecks...,
	)

	multichecker.Main(
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		testLint.ErrExitMainCheckAnalyzer,
	)

}
