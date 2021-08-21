package metadatavisitors

import (
	"fmt"
	"infraql/internal/iql/constants"
	"infraql/internal/iql/iqlmodel"
	"infraql/internal/iql/iqlutil"
	"infraql/internal/iql/metadata"
	"infraql/internal/pkg/prettyprint"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"vitess.io/vitess/go/vt/sqlparser"
)

type SchemaRequestTemplateVisitor struct {
	MaxDepth       int
	Strategy       string
	PrettyPrinter  *prettyprint.PrettyPrinter
	visitedObjects map[string]bool
	requiredOnly   bool
}

func NewSchemaRequestTemplateVisitor(maxDepth int, strategy string, prettyPrinter *prettyprint.PrettyPrinter, requiredOnly bool) *SchemaRequestTemplateVisitor {
	return &SchemaRequestTemplateVisitor{
		MaxDepth:       maxDepth,
		Strategy:       strategy,
		PrettyPrinter:  prettyPrinter,
		visitedObjects: make(map[string]bool),
		requiredOnly:   requiredOnly,
	}
}

func (sv *SchemaRequestTemplateVisitor) recordSchemaVisited(schemaKey string) {
	sv.visitedObjects[schemaKey] = true
}

func (sv *SchemaRequestTemplateVisitor) isVisited(schemaKey string, localVisited map[string]bool) bool {
	if localVisited != nil {
		if localVisited[schemaKey] {
			return true
		}
	}
	return sv.visitedObjects[schemaKey]
}

func checkAllColumnsPresent(columns sqlparser.Columns, toInclude map[string]bool) error {
	var missingColNames []string
	if columns != nil {
		for _, col := range columns {
			cName := col.GetRawVal()
			if !toInclude[cName] {
				missingColNames = append(missingColNames, cName)
			}
		}
		if len(missingColNames) > 0 {
			return fmt.Errorf("cannot find the following columns: %s", strings.Join(missingColNames, ", "))
		}
	}
	return nil
}

func getColsMap(columns sqlparser.Columns) map[string]bool {
	retVal := make(map[string]bool)
	for _, col := range columns {
		retVal[col.GetRawVal()] = true
	}
	return retVal

}

func isColIncludable(key string, columns sqlparser.Columns, colMap map[string]bool) bool {
	colOk := columns == nil
	if colOk {
		return colOk
	}
	return colMap[key]
}

func isBodyParam(paramName string) bool {
	return strings.HasPrefix(paramName, constants.RequestBodyBaseKey)
}

func ToInsertStatement(columns sqlparser.Columns, m *metadata.Method, schemaMap map[string]metadata.Schema, extended bool, prettyPrinter *prettyprint.PrettyPrinter, requiredOnly bool) (string, error) {
	paramsToInclude := m.Parameters
	successfullyIncludedCols := make(map[string]bool)
	if !extended {
		paramsToInclude = m.GetRequiredParameters()
	}
	if columns != nil {
		paramsToInclude = make(map[string]iqlmodel.Parameter)
		for _, col := range columns {
			cName := col.GetRawVal()
			if !isBodyParam(cName) {
				p, ok := m.Parameters[cName]
				if !ok {
					return "", fmt.Errorf("cannot generate insert statement: column '%s' not present", cName)
				}
				paramsToInclude[cName] = p
				successfullyIncludedCols[cName] = true
			}
		}
	}
	var includedParamNames []string
	for k, _ := range paramsToInclude {
		includedParamNames = append(includedParamNames, k)
	}
	sort.Strings(includedParamNames)
	var columnList, exprList []string
	for _, s := range includedParamNames {
		columnList = append(columnList, prettyPrinter.RenderColumnName(s))
		switch m.Parameters[s].Type {
		case "string":
			exprList = append(exprList, prettyPrinter.RenderTemplateVarAndDelimit(s))
		default:
			exprList = append(exprList, prettyPrinter.RenderTemplateVarNoDelimit(s))
		}
	}

	var sch *metadata.Schema
	if m.RequestType.Type != "" {
		s, ok := schemaMap[m.RequestType.Type]
		if ok {
			sch = &s
		}
	}

	if sch == nil {
		err := checkAllColumnsPresent(columns, successfullyIncludedCols)
		return "INSERT INTO %s" + "(\n" + strings.Join(columnList, ",\n") +
			"\n)\n" + "SELECT\n" + strings.Join(exprList, ",\n") + "\n;\n", err
	}

	schemaVisitor := NewSchemaRequestTemplateVisitor(2, "", prettyPrinter, requiredOnly)

	tVal, _ := schemaVisitor.RetrieveTemplate(sch, m, extended)

	log.Infoln(fmt.Sprintf("tVal = %v", tVal))

	colMap := getColsMap(columns)

	if columns != nil {
		for _, c := range columns {
			cName := c.GetRawVal()
			if !isBodyParam(cName) {
				continue
			}
			cNameSuffix := strings.TrimPrefix(cName, constants.RequestBodyBaseKey)
			if v, ok := tVal[cNameSuffix]; ok {
				columnList = append(columnList, prettyPrinter.RenderColumnName(cName))
				exprList = append(exprList, v)
				successfullyIncludedCols[cName] = true
			}
		}
	} else {
		tValKeysSorted := iqlutil.GetSortedKeysStringMap(tVal)
		for _, k := range tValKeysSorted {
			v := tVal[k]
			if isColIncludable(k, columns, colMap) {
				columnList = append(columnList, prettyPrinter.RenderColumnName(constants.RequestBodyBaseKey+k))
				exprList = append(exprList, v)
			}
		}
	}

	err := checkAllColumnsPresent(columns, successfullyIncludedCols)
	retVal := "INSERT INTO %s" + "(\n" + strings.Join(columnList, ",\n") +
		"\n)\n" + "SELECT\n" + strings.Join(exprList, ",\n") + "\n;\n"
	return retVal, err
}

func (sv *SchemaRequestTemplateVisitor) processSubSchemasMap(sc *metadata.Schema, method *metadata.Method, properties map[string]metadata.SchemaHandle) (map[string]string, error) {
	retVal := make(map[string]string)
	for k, v := range properties {
		ss, idStr := v.GetSchema(sc.SchemaCentral)
		log.Infoln(fmt.Sprintf("RetrieveTemplate() k = '%s', idStr = '%s' ss is nil ? '%t'", k, idStr, ss == nil))
		if ss != nil && (idStr == "" || !sv.isVisited(idStr, nil)) {
			localSchemaVisitedMap := make(map[string]bool)
			localSchemaVisitedMap[idStr] = true
			if !v.AlwaysRequired && (ss.OutputOnly || (sv.requiredOnly && !ss.IsRequired(method))) {
				log.Infoln(fmt.Sprintf("property = '%s' will be skipped", k))
				continue
			}
			rv, err := sv.retrieveTemplateVal(ss, ".values."+constants.RequestBodyBaseKey+k, localSchemaVisitedMap)
			if err != nil {
				return nil, err
			}
			switch rvt := rv.(type) {
			case map[string]interface{}, []interface{}, string:
				bytes, err := sv.PrettyPrinter.PrintTemplatedJSON(rvt)
				if err != nil {
					return nil, err
				}
				retVal[k] = string(bytes)
			case nil:
				continue
			default:
				return nil, fmt.Errorf("error processing template key '%s' with disallowed type '%T'", k, rvt)
			}
		}
	}
	return retVal, nil
}

func (sv *SchemaRequestTemplateVisitor) RetrieveTemplate(sc *metadata.Schema, method *metadata.Method, extended bool) (map[string]string, error) {
	retVal := make(map[string]string)
	var err error
	sv.recordSchemaVisited(method.RequestType.Type)
	switch sc.Type {
	case "object":
		retVal, err = sv.processSubSchemasMap(sc, method, sc.Properties)
		if len(retVal) != 0 || err != nil {
			return retVal, err
		}
		retVal, err = sv.processSubSchemasMap(sc, method, map[string]metadata.SchemaHandle{"k1": sc.AdditionalProperties})
		if len(retVal) == 0 {
			return nil, nil
		}
		return retVal, err
	}
	return nil, fmt.Errorf("templating of request body only supported for object type payload")
}

func (sv *SchemaRequestTemplateVisitor) retrieveTemplateVal(sc *metadata.Schema, objectKey string, localSchemaVisitedMap map[string]bool) (interface{}, error) {
	sSplit := strings.Split(objectKey, ".")
	oKey := sSplit[len(sSplit)-1]
	oPrefix := objectKey
	if len(sSplit) > 1 {
		oPrefix = strings.TrimSuffix(objectKey, "."+oKey)
	} else {
		oPrefix = ""
	}
	templateValSuffix := oKey
	templateValName := oPrefix + "." + templateValSuffix
	if oPrefix == "" {
		templateValName = templateValSuffix
	}
	initialLocalSchemaVisitedMap := make(map[string]bool)
	for k, v := range localSchemaVisitedMap {
		initialLocalSchemaVisitedMap[k] = v
	}
	switch sc.Type {
	case "object":
		rv := make(map[string]interface{})
		for k, v := range sc.Properties {
			propertyLocalSchemaVisitedMap := make(map[string]bool)
			for k, v := range initialLocalSchemaVisitedMap {
				propertyLocalSchemaVisitedMap[k] = v
			}
			ss, idStr := v.GetSchema(sc.SchemaCentral)
			if ss != nil && ((idStr == "" && ss.Type != "array") || !sv.isVisited(idStr, propertyLocalSchemaVisitedMap)) {
				propertyLocalSchemaVisitedMap[idStr] = true
				sv, err := sv.retrieveTemplateVal(ss, templateValName+"."+k, propertyLocalSchemaVisitedMap)
				if err != nil {
					return nil, err
				}
				if sv != nil {
					rv[k] = sv
				}
			}
		}
		if len(rv) == 0 {
			if sc.AdditionalProperties.SchemaRef != nil && len(sc.AdditionalProperties.SchemaRef) == 1 {
				for k, v := range sc.AdditionalProperties.SchemaRef {
					if k == "" {
						k = "key"
					}
					key := fmt.Sprintf("{{ %s[0].%s }}", templateValName, k)
					valBase := fmt.Sprintf("{{ %s[0].val }}", templateValName)
					switch v.Type {
					case "string":
						rv[key] = fmt.Sprintf(`"%s"`, valBase)
					case "number", "int", "int32", "int64":
						rv[key] = valBase
					default:
						rv[key] = valBase
					}
				}
			}
		}
		if len(rv) == 0 {
			return nil, nil
		}
		return rv, nil
	case "array":
		var arr []interface{}
		iSch, err := sc.GetItemsSchema()
		if err != nil {
			return nil, err
		}
		itemLocalSchemaVisitedMap := make(map[string]bool)
		for k, v := range initialLocalSchemaVisitedMap {
			itemLocalSchemaVisitedMap[k] = v
		}
		itemS, err := sv.retrieveTemplateVal(iSch, templateValName+"[0]", itemLocalSchemaVisitedMap)
		arr = append(arr, itemS)
		if err != nil {
			return nil, err
		}
		return arr, nil
	case "string":
		return "\"{{ " + templateValName + " }}\"", nil
	default:
		return "{{ " + templateValName + " }}", nil
	}
	return nil, nil
}
