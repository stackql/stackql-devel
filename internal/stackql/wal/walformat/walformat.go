package walformat

import (
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
)

type (
	WALCRC  uint32
	WALType int
)

const (
	Primitive WALType = iota
	TransactionBegin
	TransactionCommit
	TransactionRollback
)

type WALEntry interface {
	GetType() WALType
	GetCRC() WALCRC
	GetRaw() []byte
}

func NewWALEntry(
	walType WALType,
	crc WALCRC,
	tcc internaldto.TxnControlCounters, //nolint:revive // future proofing
	raw []byte,
) WALEntry {
	return newWALEntry(walType, crc, raw)
}

func newWALEntry(
	walType WALType,
	crc WALCRC,
	raw []byte,
) WALEntry {
	return &walEntry{
		walType: walType,
		crc:     crc,
		raw:     raw,
	}
}

type walEntry struct {
	walType WALType
	crc     WALCRC
	raw     []byte
}

func (w *walEntry) GetType() WALType {
	return w.walType
}

func (w *walEntry) GetCRC() WALCRC {
	return w.crc
}

func (w *walEntry) GetRaw() []byte {
	return w.raw
}
