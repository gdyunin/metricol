package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gdyunin/metricol.git/internal/server/internal/entity"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgreSQL struct {
	db *sql.DB
}

func NewPostgreSQL(connString string) (*PostgreSQL, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &PostgreSQL{db: db}, nil
}

func (p *PostgreSQL) Update(metric *entity.Metric) error {
	// TODO implement me
	panic("implement me") //nolint
}

func (p *PostgreSQL) IsExist(metricType string, metricName string) (bool, error) {
	// TODO implement me
	panic("implement me") //nolint
}

func (p *PostgreSQL) Find(metricType string, metricName string) (*entity.Metric, error) {
	// TODO implement me
	panic("implement me") //nolint
}

func (p *PostgreSQL) All() (*entity.Metrics, error) {
	// TODO implement me
	panic("implement me") //nolint
}

func (p *PostgreSQL) CheckConnection(ctx context.Context) error {
	if err := p.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

func (p *PostgreSQL) Shutdown() {
	p.close()
}

func (p *PostgreSQL) close() {
	if p.db != nil {
		_ = p.db.Close()
	}
}
