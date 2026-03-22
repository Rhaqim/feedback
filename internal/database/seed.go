package database

import (
	"context"
	"log"

	"github.com/rhaqim/worldgame/internal/game"
	"github.com/rhaqim/worldgame/internal/models"
	"github.com/rhaqim/worldgame/internal/store"
)

// Seed inserts initial regions and challenge templates into the database.
// Uses ON CONFLICT DO NOTHING / existence checks to be idempotent.
func Seed(ctx context.Context, s *store.Store) error {
	regionsSeeded := 0
	templatesSeeded := 0

	// Seed regions from the hardcoded AllRegions slice.
	for _, r := range game.AllRegions {
		if err := s.UpsertRegion(ctx, r); err != nil {
			return err
		}
		regionsSeeded++
	}

	// Seed challenge templates from the hardcoded challenge pool.
	pool := game.GetChallengePoolForSeeding()
	for tag, templates := range pool {
		// Check existing templates for this tag to avoid duplicates.
		existing, err := s.GetChallengeTemplates(ctx, tag)
		if err != nil {
			return err
		}

		existingSet := make(map[string]bool)
		for _, e := range existing {
			key := string(e.Tag) + "|" + e.TitleTemplate
			existingSet[key] = true
		}

		for _, t := range templates {
			key := string(tag) + "|" + t.Title
			if existingSet[key] {
				continue
			}

			ct := models.ChallengeTemplate{
				Tag:                 tag,
				TitleTemplate:       t.Title,
				DescriptionTemplate: t.Description,
				Source:              t.Source,
			}
			if err := s.CreateChallengeTemplate(ctx, ct); err != nil {
				return err
			}
			templatesSeeded++
		}
	}

	log.Printf("[Seed] Seeded %d regions and %d challenge templates", regionsSeeded, templatesSeeded)
	return nil
}
