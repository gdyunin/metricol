// Command staticlint is a custom static analysis tool based on golang.org/x/tools/go/analysis/multichecker.
//
// This tool statically analyzes Go code using a combination of built-in vet-style analyzers,
// select analyzers from staticcheck, go-critic, and other third-party sources.
// It is designed to be used as a custom linter executable that runs all analyzers in a single process.
//
// Usage:
//
//	go run ./cmd/staticlint ./...
//
// The tool includes the following analyzers:
//
// From golang.org/x/tools/go/analysis/passes:
//   - asmdecl: reports mismatches between assembly files and Go function declarations
//   - assign: detects useless assignments
//   - atomic: checks for common mistakes using the sync/atomic package
//   - bools: detects redundant boolean expressions
//   - buildtag: checks build constraints for correctness
//   - cgocall: detects cgo calls that may lead to performance issues
//   - composite: validates composite literal usage
//   - copylock: detects copying of values containing sync.Mutex or similar
//   - deepequalerrors: warns against using reflect.DeepEqual with errors
//   - errorsas: verifies correct use of errors.As
//   - fieldalignment: suggests struct field reordering to reduce padding
//   - httpresponse: checks that HTTP response bodies are closed
//   - loopclosure: warns about loop variable capture in goroutines or closures
//   - lostcancel: detects when context cancellation functions are not called
//   - nilfunc: checks for calls to nil functions
//   - printf: validates formatting directives in functions like fmt.Printf
//   - shadow: detects variable shadowing
//   - shift: validates bit shift operations
//   - sortslice: checks correctness of sort.Slice usage
//   - stdmethods: validates method signatures like Error() or String()
//   - stringintconv: detects suspicious string-to-int and int-to-string conversions
//   - structtag: validates struct field tags
//   - tests: checks for correct testing function signatures
//   - unmarshal: checks for issues in unmarshaling operations
//   - unreachable: detects unreachable code
//   - unsafeptr: reports invalid conversions involving unsafe.Pointer
//   - unusedresult: detects unused results of functions where the result must be used
//
// From honnef.co/go/tools/staticcheck (selected):
//   - SA (Static Analysis): a set of checks for bugs and suspicious code
//   - S1012: detect overly complex if statements
//   - ST1005: ensure error strings are lowercase
//   - QF1011: simplify redundant slice expressions
//
// From github.com/kisielk/errcheck:
//   - errcheck: reports unchecked errors from function calls
//
// From github.com/go-critic/go-critic:
//   - gocritic: a collection of optional code style and correctness checks
//
// You can customize the included analyzers by modifying the `includedStaticChecks` slice
// or the list of additional analyzers appended to `staticChecks`.
package main

import (
	"strings"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var staticChecks []*analysis.Analyzer

	includedStaticChecks := []string{
		"SA", "S1012", "ST1005", "QF1011",
	}

	for _, v := range staticcheck.Analyzers {
		for _, checkName := range includedStaticChecks {
			if strings.HasPrefix(v.Analyzer.Name, checkName) {
				staticChecks = append(staticChecks, v.Analyzer)
				break
			}
		}
	}

	staticChecks = append(staticChecks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		errcheck.Analyzer,
		gocritic.Analyzer,
	)

	multichecker.Main(staticChecks...)
}
