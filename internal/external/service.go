package external

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/rhaqim/worldgame/internal/store"
)

// FeedService manages periodic fetching and storage of feed items.
type FeedService struct {
	fetcher  *FeedFetcher
	store    *store.Store
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewFeedService creates a new FeedService that fetches on the given interval.
func NewFeedService(fetcher *FeedFetcher, s *store.Store, interval time.Duration) *FeedService {
	return &FeedService{
		fetcher:  fetcher,
		store:    s,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the background feed fetching loop. Call Stop() to shut it down.
func (fs *FeedService) Start() {
	fs.wg.Add(1)
	go fs.run()
	log.Printf("[FeedService] Started with interval %s", fs.interval)
}

// Stop gracefully shuts down the feed service.
func (fs *FeedService) Stop() {
	close(fs.stopCh)
	fs.wg.Wait()
	log.Println("[FeedService] Stopped")
}

// FetchNow runs a single fetch cycle immediately. Useful for initial population.
func (fs *FeedService) FetchNow(ctx context.Context) int {
	return fs.fetchAndStore(ctx)
}

func (fs *FeedService) run() {
	defer fs.wg.Done()

	// Do an initial fetch on startup.
	ctx := context.Background()
	fs.fetchAndStore(ctx)

	ticker := time.NewTicker(fs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fs.fetchAndStore(context.Background())
		case <-fs.stopCh:
			return
		}
	}
}

func (fs *FeedService) fetchAndStore(ctx context.Context) int {
	log.Println("[FeedService] Starting feed fetch cycle...")

	items := fs.fetcher.FetchAll(ctx)
	if len(items) == 0 {
		log.Println("[FeedService] No items fetched")
		return 0
	}

	// Enrich items: re-classify tags and detect regions.
	for i := range items {
		// Re-classify tag based on content.
		items[i].Tag = ClassifyTag(items[i].Title, items[i].Description, items[i].Tag)

		// Try to detect region from content if not already set.
		if items[i].RegionID == "" {
			items[i].RegionID = DetectRegion(items[i].Title, items[i].Description)
		}
	}

	// Store in batches.
	batchSize := 50
	stored := 0
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		if err := fs.store.UpsertFeedItems(ctx, batch); err != nil {
			log.Printf("[FeedService] Error storing batch: %v", err)
			continue
		}
		stored += len(batch)
	}

	log.Printf("[FeedService] Stored %d/%d feed items", stored, len(items))
	return stored
}
