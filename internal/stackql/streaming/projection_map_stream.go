package streaming

import (
	"io"
)

type SimpleProjectionMapStream struct {
	store      []map[string]interface{}
	projection []string
}

func NewSimpleProjectionMapStream(projection []string) MapStream {
	return &SimpleProjectionMapStream{
		projection: projection,
	}
}

func (ss *SimpleProjectionMapStream) iStackQLReader() {}

func (ss *SimpleProjectionMapStream) iStackQLWriter() {}

func (ss *SimpleProjectionMapStream) Write(input []map[string]interface{}) error {
	ss.store = append(ss.store, input...)
	return nil
}

func (ss *SimpleProjectionMapStream) Read() ([]map[string]interface{}, error) {
	var rv []map[string]interface{}
	for _, row := range ss.store {
		rowTransformed := map[string]interface{}{}
		for _, k := range ss.projection {
			v, ok := row[k]
			if ok {
				rowTransformed[k] = v
			}
		}
		rv = append(rv, rowTransformed)
	}
	return rv, io.EOF
}
