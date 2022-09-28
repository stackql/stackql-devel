package tablenamespace

import (
	"bytes"
	"regexp"
	"text/template"

	"github.com/stackql/stackql/internal/stackql/sqlengine"
)

type TableNamespaceConfigurator interface {
	GetObjectName(string) string
	IsAllowed(string) bool
	Match(string) bool
	RenderTemplate(string) (string, error)
}

var (
	_ TableNamespaceConfigurator = &regexTableNamespaceConfigurator{}
)

type regexTableNamespaceConfigurator struct {
	sqlEngine sqlengine.SQLEngine
	regex     *regexp.Regexp
	template  *template.Template
	ttl       int
}

func (stc *regexTableNamespaceConfigurator) IsAllowed(tableString string) bool {
	return stc.isAllowed(tableString)
}

func (stc *regexTableNamespaceConfigurator) isAllowed(tableString string) bool {
	return stc.regex.MatchString(tableString)
}

func (stc *regexTableNamespaceConfigurator) Match(tableString string) bool {
	isAllowed := stc.isAllowed(tableString)
	if !isAllowed {
		return false
	}
	actualTableName, err := stc.renderTemplate(tableString)
	if err != nil {
		return false
	}
	isPresent := stc.sqlEngine.IsTablePresent(actualTableName)
	if !isPresent {
		return false
	}
	// oldestUpdate := stc.sqlEngine.TableOldestUpdate(actualTableName, "iql_last_modified")
	// diff := time.Since(oldestUpdate)
	// ds := diff.Seconds()
	// if stc.ttl > 0 && int(ds) > stc.ttl {
	// 	return false
	// }
	return true
}

func (stc *regexTableNamespaceConfigurator) RenderTemplate(input string) (string, error) {
	return stc.renderTemplate(input)
}

func (stc *regexTableNamespaceConfigurator) renderTemplate(input string) (string, error) {
	objName := stc.getObjectName(input)
	inputMap := map[string]interface{}{
		"objectName": objName,
	}
	return stc.render(inputMap)
}

func (stc *regexTableNamespaceConfigurator) render(input map[string]interface{}) (string, error) {
	var tplWr bytes.Buffer
	if err := stc.template.Execute(&tplWr, input); err != nil {
		return "", err
	}
	return tplWr.String(), nil
}

func (stc *regexTableNamespaceConfigurator) GetObjectName(inputString string) string {
	return stc.getObjectName(inputString)
}

func (stc *regexTableNamespaceConfigurator) getObjectName(inputString string) string {
	for i, name := range stc.regex.SubexpNames() {
		if name == "objectName" {
			submatches := stc.regex.FindStringSubmatch(inputString)
			if len(submatches) > i {
				return submatches[i]
			}
		}
	}
	return ""
}
