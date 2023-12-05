package waltranslate

import (
	"github.com/stackql/stackql/internal/stackql/primitivegraph"
	"github.com/stackql/stackql/internal/stackql/wal/walformat"
)

var (
	_ WALTranslator = (*walTranslator)(nil)
)

type WALTranslator interface {
	Translate(walEntries []walformat.WALEntry) ([]primitivegraph.PrimitiveGraphHolder, error)
}

type walTranslator struct {
	accruedLogs []primitivegraph.PrimitiveGraphHolder
}

func (wr *walTranslator) Translate(entries []walformat.WALEntry) ([]primitivegraph.PrimitiveGraphHolder, error) {
	for _, entry := range entries {
		switch entry.GetType() { //nolint:gocritic,exhaustive // prefer this
		case walformat.Primitive:
		}
	}
	return wr.accruedLogs, nil
}
