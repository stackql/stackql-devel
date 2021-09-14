package primitive

import (
	"infraql/internal/iql/dto"
	"io"
)

type IPrimitiveCtx interface {
	GetAuthContext(string) (*dto.AuthCtx, error)
	GetWriter() io.Writer
	GetErrWriter() io.Writer
}

type IPrimitive interface {
	Optimise() error

	Execute(IPrimitiveCtx) dto.ExecutorOutput

	SetTxnId(int)

	IncidentData(int64, dto.ExecutorOutput) error

	SetInputAlias(string, int64) error
}
