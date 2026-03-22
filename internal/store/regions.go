package store

import (
	"context"
	"fmt"

	"github.com/rhaqim/worldgame/internal/models"
)

// GetAllRegions returns all regions from the database.
func (s *Store) GetAllRegions(ctx context.Context) ([]models.Region, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, country, continent, description FROM regions ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query regions: %w", err)
	}
	defer rows.Close()

	var regions []models.Region
	for rows.Next() {
		var r models.Region
		if err := rows.Scan(&r.ID, &r.Name, &r.Country, &r.Continent, &r.Description); err != nil {
			return nil, fmt.Errorf("scan region: %w", err)
		}
		regions = append(regions, r)
	}
	return regions, rows.Err()
}

// GetRegionByID returns a single region by its ID, or nil if not found.
func (s *Store) GetRegionByID(ctx context.Context, id string) (*models.Region, error) {
	var r models.Region
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, country, continent, description FROM regions WHERE id = $1`, id).
		Scan(&r.ID, &r.Name, &r.Country, &r.Continent, &r.Description)
	if err != nil {
		return nil, fmt.Errorf("get region %s: %w", id, err)
	}
	return &r, nil
}

// UpsertRegion inserts a region or does nothing if it already exists.
func (s *Store) UpsertRegion(ctx context.Context, r models.Region) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO regions (id, name, country, continent, description)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO NOTHING`,
		r.ID, r.Name, r.Country, r.Continent, r.Description)
	if err != nil {
		return fmt.Errorf("upsert region: %w", err)
	}
	return nil
}
