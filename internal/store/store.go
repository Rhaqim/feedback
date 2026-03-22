package store

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides database access for all game entities.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new Store backed by the given connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
