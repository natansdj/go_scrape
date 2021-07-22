package memory

import (
	"sync/atomic"
)

// statApp is app status structure
type statApp struct {
	TotalCount int64 `json:"total_count"`
}

// New func implements the storage interface for go_scrape (https://github.com/natansdj/go_scrape)
func New() *Storage {
	return &Storage{
		stat: &statApp{},
	}
}

// Storage is interface structure
type Storage struct {
	stat *statApp
}

// Init client storage.
func (s *Storage) Init() error {
	return nil
}

// Close the storage connection
func (s *Storage) Close() error {
	return nil
}

// Reset Client storage.
func (s *Storage) Reset() {
	atomic.StoreInt64(&s.stat.TotalCount, 0)
}

// AddTotalCount record push notification count.
func (s *Storage) AddTotalCount(count int64) {
	atomic.AddInt64(&s.stat.TotalCount, count)
}

// GetTotalCount show counts of all notification.
func (s *Storage) GetTotalCount() int64 {
	count := atomic.LoadInt64(&s.stat.TotalCount)

	return count
}
