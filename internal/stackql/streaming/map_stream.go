package streaming

import (
	"errors"
	"io"

	"github.com/stackql/stackql/pkg/maths"
)

type StackQLReader interface{}

type StackQLWriter interface{}

type StackQLReadWriter interface {
	StackQLReader
	StackQLWriter
}

type MapReader interface {
	StackQLReader
	Read() ([]map[string]interface{}, error)
}

type MapWriter interface {
	StackQLWriter
	Write([]map[string]interface{}) error
}

type StandardMapStream struct {
	store []map[string]interface{}
}

type MapStream interface {
	MapReader
	MapWriter
}

type MapStreamCollection interface {
	MapStream
	Push(MapStream)
}

func NewStandardMapStreamCollection() MapStreamCollection {
	return &standardMapStreamCollection{}
}

type standardMapStreamCollection struct {
	store   []map[string]interface{}
	streams []MapStream
}

func (sc *standardMapStreamCollection) Push(stream MapStream) {
	sc.streams = append(sc.streams, stream)
}

func (sc *standardMapStreamCollection) Write(input []map[string]interface{}) error {
	sc.store = append(sc.store, input...)
	return nil
}

func (sc *standardMapStreamCollection) Read() ([]map[string]interface{}, error) {
	var allOutputs [][]map[string]interface{}
	maxLength := 0
	var allLengths []int
	storeLen := len(sc.store)
	if storeLen > 0 {
		allLengths = append(allLengths, len(sc.store))
		maxLength = len(sc.store)
	}
	for _, stream := range sc.streams {
		output, err := stream.Read()
		if !errors.Is(err, io.EOF) {
			return output, err
		}
		thisLen := len(output)
		allOutputs = append(allOutputs, output)
		allLengths = append(allLengths, thisLen)
		if thisLen > maxLength {
			maxLength = thisLen
		}
	}
	if maxLength == 0 {
		return nil, io.EOF
	}
	lcm := maths.LcmMultiple(allLengths...)
	rv := make([]map[string]interface{}, lcm)
	for i := range rv {
		rv[i] = make(map[string]interface{})
	}
	for _, output := range allOutputs {
		thisLen := len(output)
		for i := 0; i < lcm; i++ {
			thisMap := output[i%thisLen]
			for k, v := range thisMap {
				rv[i][k] = v
			}
		}
	}
	if storeLen > 0 {
		for i := 0; i < lcm; i++ {
			thisMap := sc.store[i%storeLen]
			for k, v := range thisMap {
				rv[i][k] = v
			}
		}
	}
	return rv, io.EOF
}

func NewStandardMapStream() MapStream {
	return &StandardMapStream{}
}

func (ss *StandardMapStream) Write(input []map[string]interface{}) error {
	ss.store = append(ss.store, input...)
	return nil
}

func (ss *StandardMapStream) Read() ([]map[string]interface{}, error) {
	rv := ss.store
	ss.store = nil
	return rv, io.EOF
}
