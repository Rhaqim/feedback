package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a connection pool to the PostgreSQL database.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	log.Println("[DB] Connected to PostgreSQL")
	return pool, nil
}

// Migrate runs all CREATE TABLE IF NOT EXISTS statements to set up the schema.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
	CREATE TABLE IF NOT EXISTS regions (
		id VARCHAR(50) PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		country VARCHAR(100) NOT NULL,
		continent VARCHAR(50) NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS challenge_templates (
		id SERIAL PRIMARY KEY,
		tag VARCHAR(20) NOT NULL,
		title_template TEXT NOT NULL,
		description_template TEXT NOT NULL,
		source VARCHAR(50) NOT NULL DEFAULT 'news',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS games (
		id VARCHAR(100) PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		region_id VARCHAR(50) NOT NULL REFERENCES regions(id),
		region_name VARCHAR(100) NOT NULL,
		phase VARCHAR(20) NOT NULL DEFAULT 'active',
		week_number INTEGER NOT NULL DEFAULT 1,
		week_start TIMESTAMPTZ NOT NULL,
		week_end TIMESTAMPTZ NOT NULL,
		host_id VARCHAR(100) NOT NULL DEFAULT '',
		winner_player_id VARCHAR(100),
		winner_player_name VARCHAR(100),
		winner_score FLOAT,
		winner_summary TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS game_tags (
		game_id VARCHAR(100) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
		tag VARCHAR(20) NOT NULL,
		PRIMARY KEY (game_id, tag)
	);

	CREATE TABLE IF NOT EXISTS players (
		id VARCHAR(100) PRIMARY KEY,
		game_id VARCHAR(100) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		points FLOAT NOT NULL DEFAULT 100,
		total_score FLOAT NOT NULL DEFAULT 0,
		connected BOOLEAN NOT NULL DEFAULT false,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS challenges (
		id VARCHAR(100) PRIMARY KEY,
		game_id VARCHAR(100) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
		tag VARCHAR(20) NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		source VARCHAR(50) NOT NULL DEFAULT 'news',
		source_url TEXT NOT NULL DEFAULT '',
		region VARCHAR(100) NOT NULL,
		severity INTEGER NOT NULL DEFAULT 5,
		active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS proposals (
		id VARCHAR(100) PRIMARY KEY,
		game_id VARCHAR(100) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
		player_id VARCHAR(100) NOT NULL REFERENCES players(id),
		player_name VARCHAR(100) NOT NULL,
		challenge_id VARCHAR(100) NOT NULL REFERENCES challenges(id),
		description TEXT NOT NULL,
		points_invested FLOAT NOT NULL,
		ai_score FLOAT NOT NULL DEFAULT 0,
		ai_feedback TEXT NOT NULL DEFAULT '',
		submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS chat_messages (
		id SERIAL PRIMARY KEY,
		game_id VARCHAR(100) NOT NULL REFERENCES games(id) ON DELETE CASCADE,
		player_id VARCHAR(100) NOT NULL,
		player_name VARCHAR(100) NOT NULL,
		message TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS feed_items (
		id SERIAL PRIMARY KEY,
		tag VARCHAR(20) NOT NULL,
		region_id VARCHAR(50) NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		url TEXT NOT NULL DEFAULT '',
		source VARCHAR(50) NOT NULL DEFAULT 'rss',
		feed_name VARCHAR(100) NOT NULL DEFAULT '',
		published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		used_in_game BOOLEAN NOT NULL DEFAULT false,
		dismissed BOOLEAN NOT NULL DEFAULT false,
		UNIQUE(url, tag)
	);

	CREATE TABLE IF NOT EXISTS curated_challenges (
		id SERIAL PRIMARY KEY,
		feed_item_id INTEGER REFERENCES feed_items(id),
		tag VARCHAR(20) NOT NULL,
		region_id VARCHAR(50) NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		source VARCHAR(50) NOT NULL DEFAULT 'rss',
		source_url TEXT NOT NULL DEFAULT '',
		severity INTEGER NOT NULL DEFAULT 5,
		active BOOLEAN NOT NULL DEFAULT true,
		used_in_game BOOLEAN NOT NULL DEFAULT false,
		curator_notes TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_players_game_id ON players(game_id);
	CREATE INDEX IF NOT EXISTS idx_challenges_game_id ON challenges(game_id);
	CREATE INDEX IF NOT EXISTS idx_proposals_game_id ON proposals(game_id);
	CREATE INDEX IF NOT EXISTS idx_proposals_challenge_id ON proposals(challenge_id);
	CREATE INDEX IF NOT EXISTS idx_chat_messages_game_id ON chat_messages(game_id);
	CREATE INDEX IF NOT EXISTS idx_challenge_templates_tag ON challenge_templates(tag);
	CREATE INDEX IF NOT EXISTS idx_feed_items_tag ON feed_items(tag);
	CREATE INDEX IF NOT EXISTS idx_feed_items_region_id ON feed_items(region_id);
	CREATE INDEX IF NOT EXISTS idx_feed_items_used ON feed_items(used_in_game);
	CREATE INDEX IF NOT EXISTS idx_curated_challenges_tag ON curated_challenges(tag);
	CREATE INDEX IF NOT EXISTS idx_curated_challenges_region ON curated_challenges(region_id);
	CREATE INDEX IF NOT EXISTS idx_curated_challenges_used ON curated_challenges(used_in_game);
	`

	if _, err := pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}

	// Add dismissed column to feed_items if it doesn't exist (for existing databases).
	_, _ = pool.Exec(ctx, `ALTER TABLE feed_items ADD COLUMN IF NOT EXISTS dismissed BOOLEAN NOT NULL DEFAULT false`)

	// Add source_url column to challenges if it doesn't exist (for existing databases).
	_, _ = pool.Exec(ctx, `ALTER TABLE challenges ADD COLUMN IF NOT EXISTS source_url TEXT NOT NULL DEFAULT ''`)

	// Add source_url column to curated_challenges if it doesn't exist (for existing databases).
	_, _ = pool.Exec(ctx, `ALTER TABLE curated_challenges ADD COLUMN IF NOT EXISTS source_url TEXT NOT NULL DEFAULT ''`)

	log.Println("[DB] Schema migration complete")
	return nil
}
