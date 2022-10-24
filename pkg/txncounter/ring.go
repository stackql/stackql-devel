package txncounter

import "sync"

const (
	max32BitUnsigned uint32 = 4294967295
)

type Ring interface{}

type thirtyTwoBitModularRing struct {
	m      *sync.Mutex
	offset uint64
	max    uint64
}
