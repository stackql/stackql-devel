package vm

var (
	_ StackqlVM = (*stackqlVM)(nil)
)

type StackqlVM interface {
}

type stackqlVM struct {
}
