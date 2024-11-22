package library

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOther_Name(t *testing.T) {
	tests := []struct {
		name  string
		other Other
		want  string
	}{
		{
			"get simple name",
			Other{
				name:  "simple_name",
				value: "some",
			},
			"simple_name",
		},
		{
			"get long name",
			Other{
				name:  "loooooooooooooooooooooooooooooooooooooooooooooooong_name",
				value: "some",
			},
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
		},
		{
			"get name with numbers",
			Other{
				name:  "123_some_567",
				value: "some",
			},
			"123_some_567",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oot := tt.other.Name()
			require.Equal(t, tt.want, oot)
		})
	}
}

func TestOther_SetName(t *testing.T) {
	tests := []struct {
		name    string
		other   *Other
		newName string
		want    string
		wantErr error
	}{
		{
			"set simple name",
			&Other{},
			"simple_name",
			"simple_name",
			nil,
		},
		{
			"set long name",
			&Other{},
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
			"loooooooooooooooooooooooooooooooooooooooooooooooong_name",
			nil,
		},
		{
			"set name with numbers",
			&Other{},
			"123_some_567",
			"123_some_567",
			nil,
		},
		{
			"set empty name",
			&Other{},
			"",
			"",
			errors.New(metrics.ErrorEmptyName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.other

			err := o.SetName(tt.newName)
			if err != nil {
				require.EqualError(t, tt.wantErr, err.Error())
				return
			}

			require.Equal(t, tt.want, o.name)
		})
	}
}

func TestOther_SetValue(t *testing.T) {

	tests := []struct {
		name     string
		other    *Other
		newValue string
		want     string
		wantErr  error
	}{
		{
			"set valid int",
			&Other{},
			"5.3",
			"5.3",
			nil,
		},
		{
			"set empty value",
			&Other{},
			"",
			"0",
			errors.New(metrics.ErrorEmptyValue),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.other

			err := o.SetValue(tt.newValue)
			if err != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.Equal(t, tt.want, o.value)
		})
	}
}

func TestOther_Type(t *testing.T) {
	tests := []struct {
		name  string
		other Other
		want  metrics.MetricType
	}{
		{
			"get other type",
			Other{},
			metrics.MetricTypeOther,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oot := tt.other.Type()
			require.Equal(t, tt.want, oot)
		})
	}
}

func TestOther_Value(t *testing.T) {
	tests := []struct {
		name  string
		other Other
		want  string
	}{
		{
			"get other simple value",
			Other{
				name:  "simple_name",
				value: "some",
			},
			"some",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.other

			v := o.Value()
			require.Equal(t, tt.want, v)
		})
	}
}

func TestNewOther(t *testing.T) {
	tests := []struct {
		name string
		want *Other
	}{
		{
			"init new other",
			&Other{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOther()
			require.Equal(t, tt.want, o)
		})
	}
}
