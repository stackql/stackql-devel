package internaldto

var (
	_ FromSubtree = &fromSubtree{}
)

type FromSubtree interface {
}

type fromSubtree struct {
}

func NewFromSubtree() (FromSubtree, error) {
	return &fromSubtree{}, nil
}

func (f *fromSubtree) Render() string {
	return ""
}
