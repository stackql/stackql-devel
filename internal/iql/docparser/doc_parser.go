package docparser

import (
	"fmt"

	"infraql/internal/iql/drm"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/sqlengine"
	"infraql/internal/iql/util"

	"infraql/internal/pkg/openapistackql"

	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	SchemaDelimiter            string = "."
	googleServiceKeyDelimiter  string = ":"
	infraqlServiceKeyDelimiter string = "__"
)

var (
	drmConfig        drm.DRMConfig  = drm.GetGoogleV1SQLiteConfig()
	outputOnlyRegexp *regexp.Regexp = regexp.MustCompile(`(?i)^\[Output.Only\].*$`)
	requiredRegexpV2 *regexp.Regexp = regexp.MustCompile(`(?i)^\[Required\].*$`)
)

func TranslateServiceKeyGenericProviderToIql(serviceKey string) string {
	return strings.Replace(serviceKey, googleServiceKeyDelimiter, infraqlServiceKeyDelimiter, -1)
}

func TranslateServiceKeyIqlToGenericProvider(serviceKey string) string {
	return strings.Replace(serviceKey, infraqlServiceKeyDelimiter, googleServiceKeyDelimiter, -1)
}

func OpenapiStackQLServiceDiscoveryDocParser(bytes []byte, dbEngine sqlengine.SQLEngine, prefix string) (*openapistackql.Service, error) {
	fields := strings.Split(prefix, ".")
	if len(fields) != 2 {
		return nil, fmt.Errorf("improper resource prefix '%s'", prefix)
	}
	provStr := fields[0]
	svcStr := fields[1]
	retVal, err := openapistackql.LoadServiceDocFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	// var result map[string]interface{}
	// jsonErr := json.Unmarshal(bytes, &result)
	discoveryGenerationId, err := dbEngine.GetCurrentDiscoveryGenerationId(prefix)
	if err != nil {
		discoveryGenerationId, err = dbEngine.GetNextDiscoveryGenerationId(prefix)
		if err != nil {
			return nil, err
		}
	}
	version := retVal.Info.Version
	var tabluationsAnnotated []util.AnnotatedTabulation
	for name, s := range retVal.Components.Schemas {
		v := openapistackql.NewSchema(s.Value, name)
		if v.IsArrayRef() {
			continue
		}
		// tableName := fmt.Sprintf("%s.%s", prefix, k)
		switch v.Type {
		case "object":
			tabulation := v.Tabulate(false)
			annTab := util.NewAnnotatedTabulation(tabulation, dto.NewHeirarchyIdentifiers(provStr, svcStr, tabulation.GetName(), ""))
			tabluationsAnnotated = append(tabluationsAnnotated, annTab)
			if version == "v2" {
				for pr, prVal := range v.Properties {
					prValSc := openapistackql.NewSchema(prVal.Value, pr)
					if prValSc != nil {
						if prValSc.IsArrayRef() {
							iSc := openapistackql.NewSchema(prValSc.Items.Value, fmt.Sprintf("%s.%s.Items", v.Title, pr))
							tb := iSc.Tabulate(false)
							log.Infoln(fmt.Sprintf("tb = %v", tb))
							if tb != nil {
								annTab := util.NewAnnotatedTabulation(tb, dto.NewHeirarchyIdentifiers(provStr, svcStr, tb.GetName(), ""))
								tabluationsAnnotated = append(tabluationsAnnotated, annTab)
							}
						}
					}
				}
			}
			// create table
		case "array":
			itemsSchema, _ := v.GetItemsSchema()
			if len(itemsSchema.Properties) > 0 {
				// create "inline" table
				tabulation := v.Tabulate(false)
				annTab := util.NewAnnotatedTabulation(tabulation, dto.NewHeirarchyIdentifiers(provStr, svcStr, tabulation.GetName(), ""))
				tabluationsAnnotated = append(tabluationsAnnotated, annTab)
			}
		}
	}
	db, err := dbEngine.GetDB()
	if err != nil {
		return nil, err
	}
	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}
	for _, tblt := range tabluationsAnnotated {
		ddl := drmConfig.GenerateDDL(tblt, discoveryGenerationId)
		for _, q := range ddl {
			// log.Infoln(q)
			_, err = db.Exec(q)
			if err != nil {
				errStr := fmt.Sprintf("aborting DDL run on query = %s, err = %v", q, err)
				log.Infoln(errStr)
				txn.Rollback()
				return nil, err
			}
		}
	}
	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return retVal, nil
}

func OpenapiStackQLRootDiscoveryDocParser(bytes []byte, dbEngine sqlengine.SQLEngine, prefix string) (*openapistackql.Provider, error) {
	retVal, err := openapistackql.LoadProviderDocFromBytes(bytes)
	return retVal, err
}

func isAlwaysRequired(item interface{}) bool {
	if rMap, ok := item.(map[string]interface{}); ok {
		if desc, ok := rMap["description"]; ok {
			if descStr, ok := desc.(string); ok {
				return requiredRegexpV2.MatchString(descStr)
			}
		}
	}
	return false
}

func getRequiredIfPresent(item interface{}) map[string]bool {
	var retVal map[string]bool
	if item != nil {
		if rMap, ok := item.(map[string]interface{}); ok {
			if ref, ok := rMap["annotations"]; ok {
				if ann, ok := ref.(map[string]interface{}); ok {
					if req, ok := ann["required"]; ok {
						switch req := req.(type) {
						case []interface{}:
							retVal = make(map[string]bool)
							for _, s := range req {
								switch v := s.(type) {
								case string:
									retVal[v] = true
								}
							}
						}
					}
				}
			}
		}
	}
	return retVal
}
