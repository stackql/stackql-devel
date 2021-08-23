package requests

import (
	"encoding/json"
	"infraql/internal/iql/constants"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/metadata"
	"infraql/internal/iql/provider"
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

func SplitHttpParameters(prov provider.IProvider, sqlParamMap map[string]interface{}, method *metadata.Method, requestSchema *metadata.Schema, responseSchema *metadata.Schema) (*dto.HttpParameters, error) {
	retVal := dto.NewHttpParameters()
	for k, v := range sqlParamMap {
		var assignedToRequest bool
		if param, ok := method.Parameters[k]; ok {
			if param.Location == "query" {
				retVal.QueryParams[k] = v
				assignedToRequest = true
			} else if param.Location == "path" {
				retVal.PathParams[k] = v
				assignedToRequest = true
			}
		}
		if !assignedToRequest {
			if requestSchema != nil {
				rbp := parseRequestBodyParam(k, v)
				if rbp != nil {
					retVal.RequestBody[rbp.Key] = rbp.Val
				}
			}
		}
		if responseSchema != nil && responseSchema.FindByPath(k, nil) != nil {
			retVal.ResponseBody[k] = v
		}

	}
	return retVal, nil
}
