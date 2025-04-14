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
