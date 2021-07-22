package storage

const (
	// TotalCountKey is key name for total count of storage
	TotalCountKey = "go_scrape-total-count"
)

// Storage interface
type Storage interface {
	Init() error
	Reset()
	AddTotalCount(int64)
	GetTotalCount() int64
	Close() error
}
