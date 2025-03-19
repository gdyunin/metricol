package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"go.uber.org/zap"
)

// newTestPostgreSQL returns a repository instance with the provided sql.DB and a no‑op logger.
func newTestPostgreSQL(db *sql.DB) *PostgreSQL {
	return &PostgreSQL{
		db:     db,
		logger: zap.NewNop().Sugar(),
		dsn:    "dsn",
	}
}

func TestPostgreSQL_Update(t *testing.T) {
	tests := []struct {
		metric  *entity.Metric
		setup   func(mock sqlmock.Sqlmock)
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name:    "nil metric",
			metric:  nil,
			setup:   func(mock sqlmock.Sqlmock) {},
			wantErr: true,
			errMsg:  "metric should be non-nil",
		},
		{
			name: "JSON marshal error",
			metric: &entity.Metric{
				Type:  "gauge",
				Name:  "test",
				Value: func() {}, // functions cannot be marshaled
			},
			setup:   func(mock sqlmock.Sqlmock) {},
			wantErr: true,
			errMsg:  "failed to marshal metric value",
		},
		{
			name: "ExecContext error",
			metric: &entity.Metric{
				Type:  "counter",
				Name:  "test",
				Value: 10,
			},
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`)
				// json.Marshal(10) returns "10"
				mock.ExpectExec(query).
					WithArgs("counter", "test", []byte("10")).
					WillReturnError(errors.New("exec error"))
			},
			wantErr: true,
			errMsg:  "failed to execute query",
		},
		{
			name: "successful update",
			metric: &entity.Metric{
				Type:  "counter",
				Name:  "test",
				Value: 10,
			},
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`)
				jsonVal, _ := json.Marshal(10)
				mock.ExpectExec(query).
					WithArgs("counter", "test", jsonVal).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer func() { _ = db.Close() }()
			p := newTestPostgreSQL(db)

			tc.setup(mock)
			err = p.Update(context.Background(), tc.metric)
			if (err != nil) != tc.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil && tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("Update() error = %v, expected to contain %q", err, tc.errMsg)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgreSQL_UpdateBatch(t *testing.T) {
	tests := []struct {
		metrics *entity.Metrics
		setup   func(mock sqlmock.Sqlmock)
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name:    "nil metrics",
			metrics: nil,
			setup:   func(mock sqlmock.Sqlmock) {},
			wantErr: true,
			errMsg:  "metrics should be non-nil",
		},
		{
			name:    "contains nil metric",
			metrics: &entity.Metrics{nil},
			setup: func(mock sqlmock.Sqlmock) {
				// Expect Begin and then Rollback (since the nil metric is found after Begin)
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "metric should be non-nil",
		},
		{
			name: "tx begin error",
			metrics: &entity.Metrics{
				&entity.Metric{Type: "gauge", Name: "test", Value: 1},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			wantErr: true,
			errMsg:  "failed at begin transaction",
		},
		{
			name: "JSON marshal error",
			metrics: &entity.Metrics{
				&entity.Metric{Type: "gauge", Name: "test", Value: func() {}},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "failed to marshal metric value",
		},
		{
			name: "ExecContext error",
			metrics: &entity.Metrics{
				&entity.Metric{Type: "counter", Name: "test", Value: 5},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				query := regexp.QuoteMeta(`
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`)
				jsonVal, _ := json.Marshal(5)
				mock.ExpectExec(query).
					WithArgs("counter", "test", jsonVal).
					WillReturnError(errors.New("exec error"))
				// Rollback is triggered by the defer.
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "failed to execute query",
		},
		{
			name: "Commit error",
			metrics: &entity.Metrics{
				&entity.Metric{Type: "counter", Name: "test", Value: 5},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				query := regexp.QuoteMeta(`
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`)
				jsonVal, _ := json.Marshal(5)
				mock.ExpectExec(query).
					WithArgs("counter", "test", jsonVal).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr: true,
			errMsg:  "failed at commit transaction",
		},
		{
			name: "successful update batch",
			metrics: &entity.Metrics{
				&entity.Metric{Type: "gauge", Name: "test", Value: 3},
				&entity.Metric{Type: "counter", Name: "test2", Value: 7},
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				query := regexp.QuoteMeta(`
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`)
				jsonVal1, _ := json.Marshal(3)
				jsonVal2, _ := json.Marshal(7)
				mock.ExpectExec(query).
					WithArgs("gauge", "test", jsonVal1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(query).
					WithArgs("counter", "test2", jsonVal2).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer func() { _ = db.Close() }()
			p := newTestPostgreSQL(db)

			tc.setup(mock)
			err = p.UpdateBatch(context.Background(), tc.metrics)
			if (err != nil) != tc.wantErr {
				t.Errorf("UpdateBatch() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil && tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("UpdateBatch() error = %v, expected to contain %q", err, tc.errMsg)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgreSQL_Find(t *testing.T) {
	tests := []struct {
		setup      func(mock sqlmock.Sqlmock)
		wantMetric *entity.Metric
		name       string
		metricType string
		metricName string
		errMsg     string
		wantErr    bool
	}{
		{
			name:       "not found",
			metricType: "counter",
			metricName: "nonexistent",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
		SELECT m_name, m_type, m_value 
		FROM metrics
		WHERE m_type = $1
		  AND m_name = $2;
	`)
				// No rows returned.
				rows := sqlmock.NewRows([]string{"m_name", "m_type", "m_value"})
				mock.ExpectQuery(query).
					WithArgs("counter", "nonexistent").
					WillReturnRows(rows)
			},
			wantMetric: nil,
			wantErr:    true,
			errMsg:     "not found in repository",
		},
		{
			name:       "Query row error",
			metricType: "gauge",
			metricName: "test",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
		SELECT m_name, m_type, m_value 
		FROM metrics
		WHERE m_type = $1
		  AND m_name = $2;
	`)
				mock.ExpectQuery(query).
					WithArgs("gauge", "test").
					WillReturnError(errors.New("query error"))
			},
			wantMetric: nil,
			wantErr:    true,
			errMsg:     "failed to execute query",
		},
		{
			name:       "JSON unmarshal error",
			metricType: "gauge",
			metricName: "test",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
		SELECT m_name, m_type, m_value 
		FROM metrics
		WHERE m_type = $1
		  AND m_name = $2;
	`)
				// Return invalid JSON in the m_value column.
				rows := sqlmock.NewRows([]string{"m_name", "m_type", "m_value"}).
					AddRow("test", "gauge", []byte("invalid json"))
				mock.ExpectQuery(query).
					WithArgs("gauge", "test").
					WillReturnRows(rows)
			},
			wantMetric: nil,
			wantErr:    true,
			errMsg:     "failed to decode JSON value",
		},
		{
			name:       "successful find",
			metricType: "counter",
			metricName: "test",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta(`
		SELECT m_name, m_type, m_value 
		FROM metrics
		WHERE m_type = $1
		  AND m_name = $2;
	`)
				jsonVal, _ := json.Marshal(10)
				rows := sqlmock.NewRows([]string{"m_name", "m_type", "m_value"}).
					AddRow("test", "counter", jsonVal)
				mock.ExpectQuery(query).
					WithArgs("counter", "test").
					WillReturnRows(rows)
			},
			wantMetric: &entity.Metric{Type: "counter", Name: "test", Value: int64(10)},
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer func() { _ = db.Close() }()
			p := newTestPostgreSQL(db)

			tc.setup(mock)
			metric, err := p.Find(context.Background(), tc.metricType, tc.metricName)
			if (err != nil) != tc.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil && tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("Find() error = %v, expected to contain %q", err, tc.errMsg)
			}
			if !tc.wantErr {
				// Compare fields manually.
				if metric.Name != tc.wantMetric.Name || metric.Type != tc.wantMetric.Type ||
					fmt.Sprintf("%v", metric.Value) != fmt.Sprintf("%v", tc.wantMetric.Value) {
					t.Errorf("Find() got = %v, want %v", metric, tc.wantMetric)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgreSQL_All(t *testing.T) {
	tests := []struct {
		setup       func(mock sqlmock.Sqlmock)
		name        string
		errMsg      string
		wantMetrics entity.Metrics
		wantErr     bool
	}{
		{
			name: "query error",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta("SELECT m_name, m_type, m_value FROM metrics;")
				mock.ExpectQuery(query).WillReturnError(errors.New("query error"))
			},
			wantMetrics: nil,
			wantErr:     true,
			errMsg:      "failed to execute query",
		},
		{
			name: "row scan error",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta("SELECT m_name, m_type, m_value FROM metrics;")
				// Provide fewer columns than expected to force a scan error.
				rows := sqlmock.NewRows([]string{"m_name", "m_type"}).
					AddRow("test", "gauge")
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			wantMetrics: nil,
			wantErr:     true,
			errMsg:      "failed to process database response",
		},
		{
			name: "JSON unmarshal error",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta("SELECT m_name, m_type, m_value FROM metrics;")
				rows := sqlmock.NewRows([]string{"m_name", "m_type", "m_value"}).
					AddRow("test", "gauge", []byte("invalid json"))
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			wantMetrics: nil,
			wantErr:     true,
			errMsg:      "failed to decode JSON value",
		},
		{
			name: "successful all",
			setup: func(mock sqlmock.Sqlmock) {
				query := regexp.QuoteMeta("SELECT m_name, m_type, m_value FROM metrics;")
				jsonVal1, _ := json.Marshal(5)
				jsonVal2, _ := json.Marshal(10)
				rows := sqlmock.NewRows([]string{"m_name", "m_type", "m_value"}).
					AddRow("test1", "counter", jsonVal1).
					AddRow("test2", "gauge", jsonVal2)
				mock.ExpectQuery(query).WillReturnRows(rows)
			},
			wantMetrics: entity.Metrics{
				&entity.Metric{Type: "counter", Name: "test1", Value: int64(5)},
				&entity.Metric{Type: "gauge", Name: "test2", Value: 10},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer func() { _ = db.Close() }()
			p := newTestPostgreSQL(db)

			tc.setup(mock)
			metrics, err := p.All(context.Background())
			if (err != nil) != tc.wantErr {
				t.Errorf("All() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err != nil && tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("All() error = %v, expected to contain %q", err, tc.errMsg)
			}
			if !tc.wantErr {
				if len(*metrics) != len(tc.wantMetrics) {
					t.Errorf("All() got %d metrics, want %d", len(*metrics), len(tc.wantMetrics))
				}
				// Compare returned metrics.
				for i, m := range *metrics {
					want := tc.wantMetrics[i]
					if m.Name != want.Name || m.Type != want.Type ||
						fmt.Sprintf("%v", m.Value) != fmt.Sprintf("%v", want.Value) {
						t.Errorf("All() metric[%d] = %v, want %v", i, m, want)
					}
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgreSQL_CheckConnection(t *testing.T) {
	tests := []struct {
		setup   func(mock sqlmock.Sqlmock)
		name    string
		errMsg  string
		wantErr bool
	}{
		{
			name: "ping error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(errors.New("ping error"))
			},
			wantErr: true,
			errMsg:  "failed to ping database",
		},
		{
			name: "ping success",
			setup: func(mock sqlmock.Sqlmock) {
				// Expect a successful ping.
				mock.ExpectPing()
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Enable ping monitoring.
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer func() { _ = db.Close() }()
			p := newTestPostgreSQL(db)

			tc.setup(mock)
			err = p.CheckConnection(context.Background())
			if (err != nil) != tc.wantErr {
				t.Errorf("CheckConnection() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil && tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("CheckConnection() error = %v, expected to contain %q", err, tc.errMsg)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgreSQL_CheckConnectionWithRetry(t *testing.T) {
	tests := []struct {
		setup          func(mock sqlmock.Sqlmock)
		name           string
		errMsg         string
		attempts       int
		attemptTimeout time.Duration
		wantErr        bool
	}{
		{
			name:           "success on first attempt",
			attempts:       2,
			attemptTimeout: 100 * time.Millisecond,
			setup: func(mock sqlmock.Sqlmock) {
				// First ping succeeds.
				mock.ExpectPing()
			},
			wantErr: false,
		},
		{
			name:           "failure on all attempts",
			attempts:       2,
			attemptTimeout: 10 * time.Millisecond,
			setup: func(mock sqlmock.Sqlmock) {
				// Expect two ping attempts that fail.
				mock.ExpectPing().WillReturnError(errors.New("ping error"))
				mock.ExpectPing().WillReturnError(errors.New("ping error"))
			},
			wantErr: true,
			errMsg:  "failed to check connection to the repository",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Enable ping monitoring.
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("failed to open sqlmock: %v", err)
			}
			defer func() { _ = db.Close() }()
			p := newTestPostgreSQL(db)

			tc.setup(mock)
			err = p.CheckConnectionWithRetry(context.Background(), tc.attempts, tc.attemptTimeout)
			if (err != nil) != tc.wantErr {
				t.Errorf("CheckConnectionWithRetry() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil && tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("CheckConnectionWithRetry() error = %v, expected to contain %q", err, tc.errMsg)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPostgreSQL_Shutdown(t *testing.T) {
	// Create a sqlmock DB and shutdown the repository.
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	p := newTestPostgreSQL(db)
	p.Shutdown()

	// After Shutdown the underlying db should be closed.
	if err := p.db.Ping(); err == nil {
		t.Errorf("expected error after shutdown, got nil")
	}
}
