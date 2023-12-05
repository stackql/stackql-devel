package tsm

import (
	"sync"

	"github.com/stackql/stackql/internal/stackql/acid/wal"
	"github.com/stackql/stackql/internal/stackql/handler"
)

var (
	_ wal.WAL = &walManager{}
)

//nolint:gochecknoglobals // singleton pattern
var (
	walOnce      sync.Once
	walSingleton wal.WAL
)

type walManager struct{}

func newWALManager(_ handler.HandlerContext) (wal.WAL, error) {
	return &walManager{}, nil
}

func GetWAL(_ handler.HandlerContext) (wal.WAL, error) {
	var err error
	walOnce.Do(func() {
		if err != nil {
			return
		}
		walSingleton = &walManager{}
	})
	return walSingleton, err
}
