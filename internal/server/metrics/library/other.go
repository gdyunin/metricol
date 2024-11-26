package library

import (
	"errors"
	"github.com/gdyunin/metricol.git/internal/server/metrics"
)

type Other struct {
	name  string
	value string
}

func NewOther() *Other {
	return &Other{}
}

func (o *Other) SetName(name string) error {
	if name == "" {
		return errors.New(metrics.ErrorEmptyName)
	}
	o.name = name
	return nil
}

func (o *Other) SetValue(val string) error {
	if val == "" {
		return errors.New(metrics.ErrorEmptyValue)
	}

	o.value = val
	return nil
}

func (o Other) Name() string {
	return o.name
}

func (o Other) Value() string {
	return o.value
}

func (o Other) Type() metrics.MetricType {
	return metrics.MetricTypeOther
}
