package walcfg

import (
	"io/fs"
)

var (
	_ WALConfig = (*walConfig)(nil)
)

type WALConfig interface {
	GetWALRootDir() string
	GetWALFileMode() fs.FileMode
}

type walConfig struct {
	walRootDir  string
	walFileMode fs.FileMode
}

func NewWALConfig(
	walRootDir string,
	walFileMode fs.FileMode) WALConfig {
	return &walConfig{
		walRootDir:  walRootDir,
		walFileMode: walFileMode,
	}
}

func (wc *walConfig) GetWALRootDir() string {
	return wc.walRootDir
}

func (wc *walConfig) GetWALFileMode() fs.FileMode {
	return wc.walFileMode
}
