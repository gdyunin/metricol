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
	// Const defaultPSQLConnectionCheckTimeout specifies the timeout for a PostgreSQL connection check.
	defaultPSQLConnectionCheckTimeout = time.Second
)

var (
	// ErrQueryExecuteFailed is returned when a SQL query execution fails.
	ErrQueryExecuteFailed = errors.New("failed to execute query")
	// QueryErrFmt is the format string for wrapping query execution errors.
	QueryErrFmt = "%w: %w"
)

// PostgreSQL represents the PostgreSQL repository.
// It holds the database connection and provides methods to interact with metrics stored in the database.
type PostgreSQL struct {
	db     *sql.DB            // db is the database connection.
	logger *zap.SugaredLogger // logger is used for logging repository operations.
	dsn    string             // dsn is the Data Source Name for the PostgreSQL connection.
}

// NewPostgreSQL creates a new PostgreSQL repository instance by establishing a database connection.
// It also runs necessary migrations to ensure the database schema is up-to-date.
//
// Parameters:
//   - logger: A logger for repository operations.
//   - connString: The connection string to establish the database connection.
//
// Returns:
//   - *PostgreSQL: A pointer to the initialized PostgreSQL repository.
//   - error: An error if the database connection fails.
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

// Update inserts a new metric into the database or updates it if it already exists.
// The metric value is serialized into JSON format before storage.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metric: A pointer to the Metric to be stored.
//
// Returns:
//   - error: An error if the operation fails.
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

// UpdateBatch inserts or updates a batch of metrics in the database using a transaction.
// Each metric is serialized to JSON format prior to execution.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metrics: A pointer to the collection of Metrics to store.
//
// Returns:
//   - error: An error if the operation fails.
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
		if err = tx.Rollback(); err != nil {
			p.logger.Errorf("SQL transaction rollback failed: %v", err)
		}
	}()

	for _, m := range *metrics {
		if m == nil {
			return errors.New("metric should be non-nil, but got nil")
		}

		var mValue []byte
		mValue, err = json.Marshal(m.Value)
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
// The stored JSON value is unmarshaled into the Metric's Value field.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metricType: The type of the metric (e.g., "counter", "gauge").
//   - metricName: The name of the metric.
//
// Returns:
//   - *entity.Metric: A pointer to the retrieved Metric.
//   - error: An error if the metric is not found or retrieval fails.
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

	err = json.Unmarshal(rawValue, &m.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON value: %w", err)
	}

	return &m, nil
}

// All retrieves all metrics from the database.
// It scans each row, unmarshals the JSON value, and compiles the metrics into a collection.
//
// Parameters:
//   - ctx: The context for the operation.
//
// Returns:
//   - *entity.Metrics: A pointer to the collection of all metrics.
//   - error: An error if the retrieval fails.
func (p *PostgreSQL) All(ctx context.Context) (*entity.Metrics, error) {
	metrics := make(entity.Metrics, 0)
	query := `SELECT m_name, m_type, m_value FROM metrics;`

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(QueryErrFmt, ErrQueryExecuteFailed, err)
	}
	defer func() {
		if err = rows.Close(); err != nil {
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

// CheckConnection verifies if the database connection is alive by pinging the database.
//
// Parameters:
//   - ctx: The context for the connection check.
//
// Returns:
//   - error: An error if the database cannot be reached.
func (p *PostgreSQL) CheckConnection(ctx context.Context) error {
	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// CheckConnectionWithRetry verifies the database connection using retry logic.
// It attempts to ping the database several times before failing.
//
// Parameters:
//   - ctx: The context for the connection check.
//   - attempts: The number of retry attempts.
//   - attemptTimeout: The timeout for each connection attempt.
//
// Returns:
//   - error: An error if the connection cannot be established within the retry limit.
func (p *PostgreSQL) CheckConnectionWithRetry(
	ctx context.Context,
	attempts int,
	attemptTimeout time.Duration,
) error {
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

// mustBuild initializes the repository by checking the connection and running migrations.
// It panics if the connection check or migrations fail.
//
// Returns:
//   - *PostgreSQL: A pointer to the fully initialized PostgreSQL repository.
func (p *PostgreSQL) mustBuild() *PostgreSQL {
	var err error

	err = p.CheckConnectionWithRetry(
		context.Background(),
		defaultAttemptsDefaultCount,
		defaultPSQLConnectionCheckTimeout,
	)
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

// runMigrations applies database migrations using the embedded SQL files.
// It ensures that the necessary database tables exist.
//
// Returns:
//   - error: An error if the migrations cannot be applied.
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
