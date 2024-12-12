/*
Package storage provides an in-memory storage implementation for managing metrics.

This package defines a Repository interface for metric operations and a Store
struct that implements this interface, allowing for the storage and retrieval
of counter and gauge metrics.
*/
package storage

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gdyunin/metricol.git/internal/metrics"
)

var (
	// ErrUnknownMetricType is returned when an unknown metric type is encountered.
	ErrUnknownMetricType = errors.New("error unknown metric type")
)

// Repository defines the interface for a metric storage system.
type Repository interface {
	// PushMetric adds a new metric to the storage.
	PushMetric(metrics.Metric) error

	// GetMetric retrieves the value of a metric by its name and type.
	GetMetric(string, string) (string, error)

	// Metrics returns all stored metrics categorized by type.
	Metrics() map[string]map[string]string

	// MetricsCount returns the total number of stored metrics.
	MetricsCount() int
}

// Store is an in-memory implementation of the Repository interface.
type Store struct {
	counters map[string]int64   // Stores counter metrics
	gauges   map[string]float64 // Stores gauge metrics
}

// NewStore creates and initializes a new Store instance.
func NewStore() *Store {
	return &Store{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

// PushMetric adds a metric to the store. It supports both Counter and Gauge types.
// Returns an error if the metric type is unknown.
func (s *Store) PushMetric(metric metrics.Metric) error {
	switch m := metric.(type) {
	case *metrics.Counter:
		s.counters[m.Name] += m.Value // Increment the counter value
	case *metrics.Gauge:
		s.gauges[m.Name] = m.Value // Set the gauge value
	default:
		return fmt.Errorf("error push metric %v: %w", metric, ErrUnknownMetricType)
	}
	return nil
}

// GetMetric retrieves the value of a specified metric by its name and type.
// Returns an error if the metric name is unknown or if the metric type is invalid.
func (s *Store) GetMetric(name, metricType string) (string, error) {
	var value string
	switch metricType {
	case metrics.MetricTypeCounter:
		v, ok := s.counters[name] // Check if the counter exists
		if !ok {
			return "", fmt.Errorf("error get metric %s %s: unknown metric name", name, metricType)
		}
		value = strconv.FormatInt(v, 10)
	case metrics.MetricTypeGauge:
		v, ok := s.gauges[name] // Check if the gauge exists
		if !ok {
			return "", fmt.Errorf("error get metric %s %s: unknown metric name", name, metricType)
		}
		value = strconv.FormatFloat(v, 'g', -1, 64)
	default:
		return "", fmt.Errorf("error get metric %s %s: %w", name, metricType, ErrUnknownMetricType)
	}

	return value, nil
}

// Metrics returns all metrics stored in the store, categorized by their type.
func (s *Store) Metrics() map[string]map[string]string {
	allMetricsMap := make(map[string]map[string]string)

	allMetricsMap[metrics.MetricTypeCounter] = s.countersMap()
	allMetricsMap[metrics.MetricTypeGauge] = s.gaugesMap()

	return allMetricsMap
}

// MetricsCount returns the total number of metrics stored in both counters and gauges.
func (s *Store) MetricsCount() int {
	return len(s.gauges) + len(s.counters)
}

// countersMap converts the counters to a map of string values for easier retrieval.
func (s *Store) countersMap() map[string]string {
	m := make(map[string]string)
	for name, value := range s.counters {
		m[name] = strconv.FormatInt(value, 10)
	}
	return m
}

// gaugesMap converts the gauges to a map of string values for easier retrieval.
func (s *Store) gaugesMap() map[string]string {
	m := make(map[string]string)
	for name, value := range s.gauges {
		m[name] = strconv.FormatFloat(value, 'g', -1, 64)
	}
	return m
}
