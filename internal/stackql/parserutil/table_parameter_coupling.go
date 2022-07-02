package parserutil

import (
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"
)

type ParamSourceType int

const (
	UnknownParam ParamSourceType = iota
	WhereParam
	JoinOnParam
)

type TableParameterCoupling interface {
	AbbreviateMap(map[string]interface{}) (map[string]interface{}, error)
	Add(ColumnarReference, ParameterMetadata, ParamSourceType) error
	GetStringified() map[string]interface{}
	ReconstituteConsumedParams(map[string]interface{}) (map[string]interface{}, error)
}

type StandardTableParameterCoupling struct {
	paramMap    ParameterMap
	colMappings map[string]ColumnarReference
}

func NewTableParameterCoupling() TableParameterCoupling {
	return &StandardTableParameterCoupling{
		paramMap:    NewParameterMap(),
		colMappings: make(map[string]ColumnarReference),
	}
}

func (tpc *StandardTableParameterCoupling) Add(col ColumnarReference, val ParameterMetadata, paramType ParamSourceType) error {
	colTyped, err := NewColumnarReference(col.Value(), paramType)
	if err != nil {
		return err
	}
	err = tpc.paramMap.Set(colTyped, val)
	if err != nil {
		return err
	}
	_, ok := tpc.colMappings[col.Name()]
	if ok {
		return fmt.Errorf("parameter '%s' already present", col.Name())
	}
	tpc.colMappings[col.Name()] = col
	return nil
}

func (tpc *StandardTableParameterCoupling) GetStringified() map[string]interface{} {
	return tpc.paramMap.GetAbbreviatedStringified()
}

func (tpc *StandardTableParameterCoupling) AbbreviateMap(verboseMap map[string]interface{}) (map[string]interface{}, error) {
	return tpc.paramMap.GetAbbreviatedStringified(), nil
}

func (tpc *StandardTableParameterCoupling) ReconstituteConsumedParams(returnedMap map[string]interface{}) (map[string]interface{}, error) {
	rv := tpc.paramMap.GetStringified()
	for k, v := range returnedMap {
		key, ok := tpc.colMappings[k]
		if !ok || v == nil {
			return nil, fmt.Errorf("no reconstitution mapping for key = '%s'", k)
		}
		switch kv := key.Value().(type) {
		case *sqlparser.ColName:
			kv.Metadata = true
		}
		keyToDelete := key.String()
		_, ok = rv[keyToDelete]
		if !ok {
			return nil, fmt.Errorf("cannot process consumed params: attempt to delete non existing key")
		}
		delete(rv, keyToDelete)
	}
	return rv, nil
}
