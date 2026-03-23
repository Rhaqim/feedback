package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// UpsertFeedItem inserts a feed item, skipping duplicates by (url, tag).
func (s *Store) UpsertFeedItem(ctx context.Context, item models.FeedItem) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO feed_items (tag, region_id, title, description, url, source, feed_name, published_at, fetched_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (url, tag) DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   fetched_at = EXCLUDED.fetched_at`,
		string(item.Tag), item.RegionID, item.Title, item.Description,
		item.URL, item.Source, item.FeedName, item.PublishedAt, item.FetchedAt)
	if err != nil {
		return fmt.Errorf("upsert feed item: %w", err)
	}
	return nil
}

// UpsertFeedItems inserts multiple feed items in a transaction.
func (s *Store) UpsertFeedItems(ctx context.Context, items []models.FeedItem) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO feed_items (tag, region_id, title, description, url, source, feed_name, published_at, fetched_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 ON CONFLICT (url, tag) DO UPDATE SET
			   title = EXCLUDED.title,
			   description = EXCLUDED.description,
			   fetched_at = EXCLUDED.fetched_at`,
			string(item.Tag), item.RegionID, item.Title, item.Description,
			item.URL, item.Source, item.FeedName, item.PublishedAt, item.FetchedAt)
		if err != nil {
			return fmt.Errorf("upsert feed item %q: %w", item.Title, err)
		}
	}

	return tx.Commit(ctx)
}

// GetUnusedFeedItems returns feed items that haven't been used or dismissed,
// filtered by tag and optionally by region. Returns up to `limit` items,
// ordered by most recently published first.
func (s *Store) GetUnusedFeedItems(ctx context.Context, tag models.Tag, regionID string, limit int) ([]models.FeedItem, error) {
	query := `SELECT id, tag, region_id, title, description, url, source, feed_name, published_at, fetched_at, used_in_game, dismissed
		 FROM feed_items
		 WHERE tag = $1 AND used_in_game = false AND dismissed = false`
	args := []interface{}{string(tag)}

	if regionID != "" {
		query += ` AND (region_id = $2 OR region_id = '')`
		args = append(args, regionID)
	}

	query += ` ORDER BY published_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query unused feed items: %w", err)
	}
	defer rows.Close()

	var items []models.FeedItem
	for rows.Next() {
		var fi models.FeedItem
		if err := rows.Scan(&fi.ID, &fi.Tag, &fi.RegionID, &fi.Title, &fi.Description,
			&fi.URL, &fi.Source, &fi.FeedName, &fi.PublishedAt, &fi.FetchedAt, &fi.UsedInGame, &fi.Dismissed); err != nil {
			return nil, fmt.Errorf("scan feed item: %w", err)
		}
		items = append(items, fi)
	}
	return items, rows.Err()
}

// MarkFeedItemUsed sets used_in_game = true for a feed item.
func (s *Store) MarkFeedItemUsed(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE feed_items SET used_in_game = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark feed item used: %w", err)
	}
	return nil
}

// DismissFeedItem marks a feed item as dismissed (not relevant).
func (s *Store) DismissFeedItem(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE feed_items SET dismissed = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("dismiss feed item: %w", err)
	}
	return nil
}

// GetRecentFeedItems returns the most recently fetched feed items.
func (s *Store) GetRecentFeedItems(ctx context.Context, limit int) ([]models.FeedItem, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, tag, region_id, title, description, url, source, feed_name, published_at, fetched_at, used_in_game, dismissed
		 FROM feed_items ORDER BY fetched_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent feed items: %w", err)
	}
	defer rows.Close()

	var items []models.FeedItem
	for rows.Next() {
		var fi models.FeedItem
		if err := rows.Scan(&fi.ID, &fi.Tag, &fi.RegionID, &fi.Title, &fi.Description,
			&fi.URL, &fi.Source, &fi.FeedName, &fi.PublishedAt, &fi.FetchedAt, &fi.UsedInGame, &fi.Dismissed); err != nil {
			return nil, fmt.Errorf("scan feed item: %w", err)
		}
		items = append(items, fi)
	}
	return items, rows.Err()
}

// GetFeedItemsFiltered returns feed items with optional tag/region/unused filters.
// Excludes dismissed items by default.
func (s *Store) GetFeedItemsFiltered(ctx context.Context, tag string, regionID string, unusedOnly bool, limit int) ([]models.FeedItem, error) {
	query := `SELECT id, tag, region_id, title, description, url, source, feed_name, published_at, fetched_at, used_in_game, dismissed
		 FROM feed_items WHERE dismissed = false`
	args := []interface{}{}
	argIdx := 1

	if tag != "" {
		query += fmt.Sprintf(" AND tag = $%d", argIdx)
		args = append(args, tag)
		argIdx++
	}
	if regionID != "" {
		query += fmt.Sprintf(" AND (region_id = $%d OR region_id = '')", argIdx)
		args = append(args, regionID)
		argIdx++
	}
	if unusedOnly {
		query += " AND used_in_game = false"
	}

	query += fmt.Sprintf(" ORDER BY published_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query filtered feed items: %w", err)
	}
	defer rows.Close()

	var items []models.FeedItem
	for rows.Next() {
		var fi models.FeedItem
		if err := rows.Scan(&fi.ID, &fi.Tag, &fi.RegionID, &fi.Title, &fi.Description,
			&fi.URL, &fi.Source, &fi.FeedName, &fi.PublishedAt, &fi.FetchedAt, &fi.UsedInGame, &fi.Dismissed); err != nil {
			return nil, fmt.Errorf("scan feed item: %w", err)
		}
		items = append(items, fi)
	}
	return items, rows.Err()
}
