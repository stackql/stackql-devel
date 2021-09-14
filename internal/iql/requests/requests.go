package requests

import (
	"encoding/json"
	"infraql/internal/iql/constants"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/metadata"
	"infraql/internal/iql/provider"
	"sort"
	"strings"
)

type requestBodyParam struct {
	Key string
	Val interface{}
}

func parseRequestBodyParam(k string, v interface{}) *requestBodyParam {
	trimmedKey := strings.TrimPrefix(k, constants.RequestBodyBaseKey)
	var parsedVal interface{}
	if trimmedKey != k {
		switch vt := v.(type) {
		case string:
			var js map[string]interface{}
			var jArr []interface{}
			if json.Unmarshal([]byte(vt), &js) == nil {
				parsedVal = js
			} else if json.Unmarshal([]byte(vt), &jArr) == nil {
				parsedVal = jArr
			} else {
				parsedVal = vt
			}
		default:
			parsedVal = vt
		}
		return &requestBodyParam{
			Key: trimmedKey,
			Val: parsedVal,
		}
	}
	return nil
}

func SplitHttpParameters(prov provider.IProvider, sqlParamMap map[int]map[string]interface{}, method *metadata.Method, requestSchema *metadata.Schema, responseSchema *metadata.Schema) ([]*dto.HttpParameters, error) {
	var retVal []*dto.HttpParameters
	var rowKeys []int
	for idx, _ := range sqlParamMap {
		rowKeys = append(rowKeys, idx)
	}
	sort.Ints(rowKeys)
	for _, k := range rowKeys {
		sqlRow := sqlParamMap[k]
		reqMap := dto.NewHttpParameters()
		for k, v := range sqlRow {
			var assignedToRequest bool
			if param, ok := method.Parameters[k]; ok {
				if param.Location == "query" {
					reqMap.QueryParams[k] = v
					assignedToRequest = true
				} else if param.Location == "path" {
					reqMap.PathParams[k] = v
					assignedToRequest = true
				}
			}
			if !assignedToRequest {
				if requestSchema != nil {
					rbp := parseRequestBodyParam(k, v)
					if rbp != nil {
						reqMap.RequestBody[rbp.Key] = rbp.Val
					}
				}
			}
			if responseSchema != nil && responseSchema.FindByPath(k, nil) != nil {
				reqMap.ResponseBody[k] = v
			}
		}
		retVal = append(retVal, reqMap)
	}
	return retVal, nil
}
