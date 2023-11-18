package walwrite

import (
	"fmt"
	"os"

	"github.com/stackql/stackql/internal/stackql/wal/walcfg"
	"github.com/stackql/stackql/internal/stackql/wal/walformat"
	"github.com/stackql/stackql/internal/stackql/wal/walstate"
)

var (
	_ WALWriter = (*walWriter)(nil)
)

type WALWriter interface {
	Write(walEntries []walformat.WALEntry) error
}

type walWriter struct {
	cfg      walcfg.WALConfig
	walState walstate.WALState
}

func (ww *walWriter) getFileName(walID walstate.WALID) string {
	pathSep := string(os.PathSeparator)
	return fmt.Sprintf("%s%s%d.wal", ww.cfg.GetWALRootDir(), pathSep, walID)
}

func (ww *walWriter) Write(entries []walformat.WALEntry) error {
	for _, entry := range entries {
		walID := ww.walState.NextWALID()
		fileName := ww.getFileName(walID)
		err := os.WriteFile(fileName, entry.GetRaw(), ww.cfg.GetWALFileMode())
		if err != nil {
			return err
		}
	}
	return nil
}
