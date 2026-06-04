/*
Package staticlint provides a multichecker for static analysis of Go code.

This tool combines multiple static analyzers into a single executable:
  - Standard Go analyzers from golang.org/x/tools/go/analysis/passes
  - All SA class analyzers from staticcheck.io (Staticcheck)
  - Additional analyzers from other staticcheck classes (S, ST, QF)
  - Custom analyzer to prohibit os.Exit calls in main functions

# Usage

To run static analysis on your project:

	go build ./cmd/staticlint
	./staticlint ./...

# Analyzers

## Standard Go Analyzers

- **printf**: Checks consistency of Printf format strings and arguments
- **shadow**: Checks for shadowed variables
- **structtag**: Checks that struct field tags conform to reflect.StructTag.Get
- **atomic**: Check for common mistakes using the sync/atomic package
- **bools**: Detects common mistakes involving boolean operators
- **composites**: Checks for unkeyed composite literals
- **copylocks**: Checks for locks erroneously passed by value
- **httpresponse**: Checks for mistakes using HTTP responses
- **loopclosure**: Checks that loop variables captured by closures are used correctly
- nilfunc: Checks for useless comparisons between functions and nil
- stdmethods: Checks signature of methods of well-known interfaces
- tests: Checks for common mistaken usages of tests and examples

## Staticcheck SA Analyzers

All SA (Static Analysis) class analyzers from staticcheck.io are included.
These check for:
- Static correctness issues (SA1000-SA1099)
- Correct usage of flags (SA2000-SA2009)
- String usage errors (SA3000-SA3012)
- Incorrect concurrency patterns (SA4000-SA4032)
- Performance issues (SA5000-SA5012)
- Deprecated functions (SA6000-SA6002)
- Issues with specific packages (SA8000-SA9014)

## Additional Staticcheck Analyzers

- **ST1000-ST1023**: Style checkers for naming conventions
- **QF1000-QF1001**: Quickfix suggestions for code simplification

## Custom Analyzers

  - **noexit**: Prohibits direct calls to os.Exit in the main function of package main.
    This encourages proper error handling and graceful shutdown.
    Use logger.Fatal or os.Exit in cleanup functions instead.

# Configuration

The analyzers run with their default configurations. For custom analysis
options, refer to the documentation of each analyzer.

# Requirements

The tool requires Go 1.21 or later. Dependencies:
- golang.org/x/tools/go/analysis/multichecker
- golang.org/x/tools/go/analysis/passes
- honnef.co/go/tools/staticcheck
*/
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"shorturl/cmd/staticlint/analyzers/noexit"
)

func main() {
	// Combine all analyzers
	analyzers := []*analysis.Analyzer{
		// Standard Go analyzers
		atomic.Analyzer,       // sync/atomic usage
		bools.Analyzer,        // boolean operator mistakes
		composite.Analyzer,    // unkeyed composite literals
		copylock.Analyzer,     // locks passed by value
		httpresponse.Analyzer, // HTTP response mistakes
		loopclosure.Analyzer,  // loop variable capture
		nilfunc.Analyzer,      // nil function comparisons
		printf.Analyzer,       // printf format strings
		shadow.Analyzer,       // shadowed variables
		stdmethods.Analyzer,   // standard interface methods
		structtag.Analyzer,    // struct tag format
		tests.Analyzer,        // test code mistakes

		// Custom analyzer
		noexit.Analyzer, // prohibit os.Exit in main function
	}

	// Add all Staticcheck SA analyzers (static analysis)
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// Add all ST analyzers (style checkers)
	for _, v := range stylecheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// Add all S analyzers (simple checks)
	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	// Add QF analyzers (quickfix suggestions)
	for _, v := range quickfix.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	multichecker.Main(analyzers...)
}
