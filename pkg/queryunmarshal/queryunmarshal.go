package queryunmarshal

type QueryUnmarshaller interface {
	Unmarshal(input interface{}) (map[string]string, error)
}
