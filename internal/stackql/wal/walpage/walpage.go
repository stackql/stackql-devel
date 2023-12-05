package walpage

var (
	_ WALPage = (*walPage)(nil)
)

type WALPage interface {
	Write() error
}

type walPage struct {
	// data []byte
}

func (wp *walPage) Write() error {
	return nil
}
