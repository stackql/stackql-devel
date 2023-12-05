package walread

import (
	"os"

	"github.com/stackql/stackql/internal/stackql/wal/walcfg"
	"github.com/stackql/stackql/internal/stackql/wal/walstate"
)

type WALReader interface {
	Read(fileName string) ([]byte, error)
}

type walReader struct {
	cfg      walcfg.WALConfig  //nolint:unused // future proofing
	walState walstate.WALState //nolint:unused // future proofing
}

func NewWALReader() WALReader {
	return &walReader{}
}

func (wr *walReader) Read(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}
