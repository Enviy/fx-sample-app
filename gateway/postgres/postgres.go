package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"go.uber.org/config"
)

// Gateway defines the methods for interacting with postgres.
type Gateway interface {
	MergeRecord(ctx context.Context, p RecordParam) (string, error)
	CreateRecord(ctx context.Context, p RecordParam) error
	UpdateRecord(ctx context.Context, p RecordParam) error
}
type gateway struct {
	db     *sql.DB
	txOpts *sql.TxOptions
}

// New is the constructor for the postgres interface.
func New(cfg config.Provider) (Gateway, error) {
	// Connection string params in config supports different envs.
	connStr := fmt.Sprintf(
		"user=%s dbname=%s password=%s sslmode=%s",
		cfg.Get("postgres.user").String(),
		cfg.Get("postgres.db_name").String(),
		cfg.Get("postgres.password").String(),
		cfg.Get("postgres.ssl_mode").String(),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("sql Open %w", err)
	}

	return &gateway{
		db: db,
		// This isolation level supports easier concurrent ops.
		txOpts: &sql.TxOptions{Isolation: sql.LevelSerializable},
	}, nil
}

// MergeRecord attempts a create transaction and a merge transaction on collision.
func (g *gateway) MergeRecord(ctx context.Context, p RecordParam) (string, error) {
	err := g.CreateRecord(ctx, p)
	if err != nil {
		// Catch primary key collisions and call MergeRecord.
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			updateErr := g.UpdateRecord(ctx, p)
			if updateErr != nil {
				return "", fmt.Errorf("UpdateRecord %w", updateErr)
			}
			// Merge happy path exit.
			return MERGED, nil
		}
		return "", fmt.Errorf("CreateRecord %w", err)
	}
	// Create happy path exit.
	return CREATED, nil
}

// CreateRecord attempts to create a new record.
func (g *gateway) CreateRecord(ctx context.Context, p RecordParam) error {
	tx, err := g.db.BeginTx(ctx, g.txOpts)
	if err != nil {
		return fmt.Errorf("BeginTx %w", err)
	}
	// Rollback is ignored if the transaction is already committed.
	defer tx.Rollback()

	// insert
	resp, err := g.db.ExecContext(
		ctx,
		insertQuery,
		p.FindingID,
		p.DetectionName,
		p.CollisionSlug,
		p.FirstEvent,
		p.LastEvent,
		pq.Array(p.RawEvents),
	)
	if err != nil {
		return fmt.Errorf("ExecContext %w", err)
	}

	fmt.Println("resp", resp)

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Commit %w", err)
	}

	return nil
}

// UpdateRecord attempts to merge raw_events into an existing record.
func (g *gateway) UpdateRecord(ctx context.Context, p RecordParam) error {
	tx, err := g.db.BeginTx(ctx, g.txOpts)
	if err != nil {
		return fmt.Errorf("BeginTx %w", err)
	}
	defer tx.Rollback()

	_, err = g.db.ExecContext(
		ctx,
		mergeQuery,
		pq.Array(p.RawEvents),
		p.CollisionSlug,
		p.FirstEvent,
		p.LastEvent,
	)
	if err != nil {
		return fmt.Errorf("ExecContext %w", err)
	}

	return nil
}
