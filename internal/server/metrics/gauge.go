package metrics

type Gauge struct {
	name  string
	value float64
}

func NewGauge(name string, value float64) *Gauge {
	return &Gauge{
		name,
		value,
	}
}

func (g Gauge) MetricName() string {
	return g.name
}

func (g Gauge) MetricValue() float64 {
	return g.value
}
