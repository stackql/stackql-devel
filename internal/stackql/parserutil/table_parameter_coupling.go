package parserutil

import (
	"fmt"

	"vitess.io/vitess/go/vt/sqlparser"
)

type TableParameterCoupling struct {
	paramMap    ParameterMap
	colMappings map[string]ColumnarReference
}

func NewTableParameterCoupling() *TableParameterCoupling {
	return &TableParameterCoupling{
		paramMap:    NewParameterMap(),
		colMappings: make(map[string]ColumnarReference),
	}
}

func (tpc *TableParameterCoupling) Add(col ColumnarReference, val ParameterMetadata) error {
	err := tpc.paramMap.Set(col, val)
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

func (tpc *TableParameterCoupling) GetStringified() map[string]interface{} {
	return tpc.paramMap.GetAbbreviatedStringified()
}

func (tpc *TableParameterCoupling) AbbreviateMap(verboseMap map[string]interface{}) (map[string]interface{}, error) {
	return tpc.paramMap.GetAbbreviatedStringified(), nil
}

func (tpc *TableParameterCoupling) ReconstituteConsumedParams(returnedMap map[string]interface{}) (map[string]interface{}, error) {
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
