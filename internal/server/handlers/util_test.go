package handlers

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_splitURL(t *testing.T) {
	type args struct {
		s string
		n int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"full fill uri",
			args{
				s: "testMetric/testName/testValue",
				n: 3,
			},
			[]string{"testMetric", "testName", "testValue"},
		},
		{
			"partially fill uri: only metric",
			args{
				s: "testMetric/",
				n: 3,
			},
			[]string{"testMetric", "", ""},
		},
		{
			"partially fill uri: metric and uri",
			args{
				s: "testMetric/testName",
				n: 3,
			},
			[]string{"testMetric", "testName", ""},
		},
		{
			"empty uri",
			args{
				s: "",
				n: 3,
			},
			[]string{"", "", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitURI(tt.args.s, tt.args.n)
			require.Equal(t, tt.want, got)
		})
	}
}
