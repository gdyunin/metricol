package repository

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	"github.com/gdyunin/metricol.git/pkg/retry"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

const (
	// PSQLDefaultConnectionCheckTimeout specifies the duration to wait for a connection check before timing out.
	defaultPSQLConnectionCheckTimeout = time.Second
	defaultPSQLCreateTablesTimeout    = 3 * time.Second
)

var (
	// ErrQueryExecuteFailed is a predefined error for handling failures during query execution.
	// Use this error to provide consistent error messaging for query-related issues.
	ErrQueryExecuteFailed = errors.New("failed to execute query")

	// QueryErrFmt is a format string used for wrapping errors related to query execution.
	// It combines multiple errors into a single error message using the fmt.Errorf function
	// and helps maintain consistency in error handling.
	QueryErrFmt = "%w: %w"
)

// PostgreSQL represents the PostgreSQL repository. It holds the database connection and provides methods
// to interact with the metrics data stored in the PostgreSQL database.
type PostgreSQL struct {
	db     *sql.DB
	logger *zap.SugaredLogger
	dsn    string
}

// NewPostgreSQL creates a new PostgreSQL repository instance by establishing a connection to the database
// using the provided connection string. It also ensures that the necessary database structure is in place.
//
// Parameters:
//   - connString: the connection string to establish the database connection.
//   - logger: Logger for repository operations.
//
// Returns:
//   - *PostgreSQL: the initialized repository instance.
//   - error: an error if the database connection fails.
func NewPostgreSQL(logger *zap.SugaredLogger, connString string) (*PostgreSQL, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	psql := PostgreSQL{
		db:     db,
		dsn:    connString,
		logger: logger,
	}
	return psql.mustBuild(), nil
}

// Update inserts a new metric into the database or updates the existing one if a metric with the same
// type and name already exists. The metric value is serialized into JSON format before being stored.
//
// Parameters:
//   - metric: the metric object containing type, name, and value to be stored.
//
// Returns:
//   - error: an error if the query execution or JSON marshaling fails.
func (p *PostgreSQL) Update(ctx context.Context, metric *entity.Metric) error {
	if metric == nil {
		return errors.New("metric should be non-nil, but got nil")
	}

	query := `
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`

	mValue, err := json.Marshal(metric.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal metric value: %w", err)
	}
	_, err = p.db.ExecContext(ctx, query, metric.Type, metric.Name, mValue)
	if err != nil {
		return fmt.Errorf(QueryErrFmt, ErrQueryExecuteFailed, err)
	}

	return nil
}

func (p *PostgreSQL) UpdateBatch(ctx context.Context, metrics *entity.Metrics) error {
	if metrics == nil {
		return errors.New("metrics should be non-nil, but got nil")
	}

	query := `
		INSERT INTO public.metrics (m_type, m_name, m_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (m_type, m_name)
		DO UPDATE SET m_value = EXCLUDED.m_value;
	`

	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed at begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			p.logger.Errorf("SQL transaction rollback failed: %v", err)
		}
	}()

	for _, m := range *metrics {
		if m == nil {
			return errors.New("metric should be non-nil, but got nil")
		}

		mValue, err := json.Marshal(m.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal metric value: %w", err)
		}

		_, err = tx.ExecContext(ctx, query, m.Type, m.Name, mValue)
		if err != nil {
			return fmt.Errorf(QueryErrFmt, ErrQueryExecuteFailed, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed at commit transaction: %w", err)
	}
	return nil
}

// Find retrieves a metric from the database based on its type and name.
//
// Parameters:
//   - metricType: the type of the metric to retrieve (e.g., "counter", "gauge").
//   - metricName: the name of the metric to retrieve.
//
// Returns:
//   - *entity.Metric: the retrieved metric object.
//   - error: an error if the metric is not found or the query execution fails.
func (p *PostgreSQL) Find(ctx context.Context, metricType string, metricName string) (*entity.Metric, error) {
	query := `
		SELECT m_name, m_type, m_value 
		FROM metrics
		WHERE m_type = $1
		  AND m_name = $2;
	`

	m := entity.Metric{}
	var rawValue []byte

	err := p.db.QueryRowContext(ctx, query, metricType, metricName).Scan(&m.Name, &m.Type, &rawValue)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: type=%s, name=%s", ErrNotFoundInRepo, metricType, metricName)
		}
		return nil, fmt.Errorf(QueryErrFmt, ErrQueryExecuteFailed, err)
	}

	// [ДЛЯ РЕВЬЮ]: Храним значение как JSONB. Подробнее в комментах к func (p *PostgreSQL) runMigrations() error.
	err = json.Unmarshal(rawValue, &m.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON value: %w", err)
	}

	return &m, nil
}

// All retrieves all metrics from the database.
//
// Returns:
//   - *entity.Metrics: a collection of all metrics stored in the database.
//   - error: an error if the query execution or row processing fails.
func (p *PostgreSQL) All(ctx context.Context) (*entity.Metrics, error) {
	metrics := make(entity.Metrics, 0)
	query := `SELECT m_name, m_type, m_value FROM metrics;`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(QueryErrFmt, ErrQueryExecuteFailed, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Errorf("SQL rows result close error: %v", err)
		}
	}()

	for rows.Next() {
		m := entity.Metric{}
		var rawValue []byte

		err = rows.Scan(&m.Name, &m.Type, &rawValue)
		if err != nil {
			return nil, fmt.Errorf("failed to process database response: %w", err)
		}

		// [ДЛЯ РЕВЬЮ]: Храним значение как JSONB. Подробнее в комментах к func (p *PostgreSQL) runMigrations() error.
		err = json.Unmarshal(rawValue, &m.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON value: %w", err)
		}

		metrics = append(metrics, &m)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to process database response: %w", err)
	}
	return &metrics, nil
}

// CheckConnection verifies if the database connection is alive by sending a ping request.
//
// Parameters:
//   - ctx: the context for managing the connection lifecycle.
//
// Returns:
//   - error: an error if the database cannot be reached.
func (p *PostgreSQL) CheckConnection(ctx context.Context) error {
	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// CheckConnectionWithRetry verifies the database connection using retry logic.
//
// Parameters:
//   - ctx: the context for managing the connection lifecycle.
//   - attempts: the number of retry attempts.
//   - attemptTimeout: the timeout for each connection attempt.
//
// Returns:
//   - error: an error if the connection cannot be established within the retry limit.
func (p *PostgreSQL) CheckConnectionWithRetry(ctx context.Context, attempts int, attemptTimeout time.Duration) error {
	if err := retry.WithRetry(ctx, p.logger, "check connection to postgre db", attempts, func() error {
		checkCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		defer cancel()
		if err := p.CheckConnection(checkCtx); err != nil {
			return fmt.Errorf("failed to check connection to the repository: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to check connection to the repository: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the repository by closing the database connection.
func (p *PostgreSQL) Shutdown() {
	p.close()
}

// close closes the database connection if it is not already closed.
func (p *PostgreSQL) close() {
	if p.db != nil {
		_ = p.db.Close()
	}
}

// mustBuild initializes the repository and ensures the database is ready for operations.
//
// Returns:
//   - *PostgreSQL: the fully initialized repository instance.
//
// Panics:
//   - If the database connection cannot be established or tables cannot be created.
func (p *PostgreSQL) mustBuild() *PostgreSQL {
	var err error

	err = p.CheckConnectionWithRetry(context.Background(), defaultAttemptsDefaultCount, defaultPSQLConnectionCheckTimeout)
	if err != nil {
		panic(fmt.Sprintf("failed to check connection to the repository: %v", err))
	}

	err = p.runMigrations()
	if err != nil {
		panic(fmt.Sprintf("failed to execute migrations: %v", err))
	}

	return p
}

//go:embed migrations/psql/*.sql
var migrationsDir embed.FS

// runMigrations ensures that the necessary database tables exist. If the tables are missing, it creates them.
//
// Returns:
//   - error: an error if the table creation query fails.
func (p *PostgreSQL) runMigrations() error {
	d, err := iofs.New(migrationsDir, "migrations/psql")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, p.dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
