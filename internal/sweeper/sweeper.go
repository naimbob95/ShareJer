// Package sweeper periodically removes expired files. This is the active
// counterpart to the lazy expiry check in the handlers: lazy deletion only
// triggers when an expired file is accessed, so without a sweep, expired-but-
// untouched files would linger on disk and in the database forever.
package sweeper

import (
	"context"
	"log"
	"time"

	"github.com/naimbob95/sharejer/internal/db"
	"github.com/naimbob95/sharejer/internal/storage"
)

// Start launches a background goroutine that sweeps every `interval`, deleting
// expired files (bytes on disk + their DB rows). It runs one sweep immediately,
// then on each tick, and stops when ctx is cancelled.
func Start(ctx context.Context, database *db.DB, store storage.Storage, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		sweep(database, store) // clean up anything already expired at startup

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sweep(database, store)
			}
		}
	}()
}

func sweep(database *db.DB, store storage.Storage) {
	expired, err := database.ExpiredFiles(time.Now())
	if err != nil {
		log.Printf("sweeper: failed to query expired files: %v", err)
		return
	}

	for _, f := range expired {
		if err := store.Delete(f.StoragePath); err != nil {
			log.Printf("sweeper: failed to delete %s from disk: %v", f.ID, err)
		}
		if err := database.DeleteFile(f.ID); err != nil {
			log.Printf("sweeper: failed to delete record %s: %v", f.ID, err)
		}
	}

	if len(expired) > 0 {
		log.Printf("sweeper: removed %d expired file(s)", len(expired))
	}
}
