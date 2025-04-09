package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

func Test_staticChecks(t *testing.T) {
	got := staticChecks()

	gotNames := extractNames(got)
	wantNames := []string{
		"SA1000", "SA1001", "SA1002", "SA1003", "SA1004", "SA1005", "SA1006", "SA1007",
		"SA1008", "SA1010", "SA1011", "SA1012", "SA1013", "SA1014", "SA1015", "SA1016",
		"SA1017", "SA1018", "SA1019", "SA1020", "SA1021", "SA1023", "SA1024", "SA1025",
		"SA1026", "SA1027", "SA1028", "SA1029", "SA1030", "SA1031", "SA1032", "SA2000",
		"SA2001", "SA2002", "SA2003", "SA3000", "SA3001", "SA4000", "SA4001", "SA4003",
		"SA4004", "SA4005", "SA4006", "SA4008", "SA4009", "SA4010", "SA4011", "SA4012",
		"SA4013", "SA4014", "SA4015", "SA4016", "SA4017", "SA4018", "SA4019", "SA4020",
		"SA4021", "SA4022", "SA4023", "SA4024", "SA4025", "SA4026", "SA4027", "SA4028",
		"SA4029", "SA4030", "SA4031", "SA4032", "SA5000", "SA5001", "SA5002", "SA5003",
		"SA5004", "SA5005", "SA5007", "SA5008", "SA5009", "SA5010", "SA5011", "SA5012",
		"SA6000", "SA6001", "SA6002", "SA6003", "SA6005", "SA6006", "SA9001", "SA9002",
		"SA9003", "SA9004", "SA9005", "SA9006", "SA9007", "SA9008", "SA9009", "QF1001",
		"QF1002", "QF1003", "QF1004", "QF1005", "QF1006", "QF1007", "QF1008", "QF1009",
		"QF1010", "QF1011", "QF1012", "S1000", "S1001", "S1002", "S1003", "S1004",
		"S1005", "S1006", "S1007", "S1008", "S1009", "S1010", "S1011", "S1012",
		"S1016", "S1017", "S1018", "S1019", "S1020", "S1021", "S1023", "S1024",
		"S1025", "S1028", "S1029", "S1030", "S1031", "S1032", "S1033", "S1034",
		"S1035", "S1036", "S1037", "S1038", "S1039", "S1040",
	}

	assert.Equal(t, wantNames, gotNames)
}

func Test_otherChecks(t *testing.T) {
	got := otherChecks()

	gotNames := extractNames(got)
	wantNames := []string{
		"errcheck",
		"ruleguard",
	}

	assert.Equal(t, wantNames, gotNames)
}

func Test_customChecks(t *testing.T) {
	got := customChecks()

	gotNames := extractNames(got)
	wantNames := []string{
		"exitcheck",
	}

	assert.Equal(t, wantNames, gotNames)
}

func Test_analysisChecks(t *testing.T) {
	got := analysisChecks()

	gotNames := extractNames(got)
	wantNames := []string{
		"asmdecl",
		"assign",
		"atomic",
		"bools",
		"buildtag",
		"cgocall",
		"composites",
		"copylocks",
		"deepequalerrors",
		"errorsas",
		"fieldalignment",
		"httpresponse",
		"loopclosure",
		"lostcancel",
		"nilfunc",
		"printf",
		"shadow",
		"shift",
		"sortslice",
		"stdmethods",
		"stringintconv",
		"structtag",
		"tests",
		"unmarshal",
		"unreachable",
		"unsafeptr",
		"unusedresult",
	}

	assert.Equal(t, wantNames, gotNames)
}

func Test_makeAlalyzersSlice(t *testing.T) {
	got := makeAlalyzersSlice()

	want := make([]*analysis.Analyzer, 0)
	want = append(want, staticChecks()...)
	want = append(want, otherChecks()...)
	want = append(want, analysisChecks()...)
	want = append(want, customChecks()...)

	gotNames := extractNames(got)
	wantNames := extractNames(want)

	assert.Equal(t, wantNames, gotNames)
}

func extractNames(analyzers []*analysis.Analyzer) []string {
	names := make([]string, 0, len(analyzers))
	for _, a := range analyzers {
		names = append(names, a.Name)
	}
	return names
}
