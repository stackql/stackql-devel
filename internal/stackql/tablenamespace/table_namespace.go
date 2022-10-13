package tablenamespace

import (
	"bytes"
	"regexp"
	"text/template"
	"time"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/sqlengine"
)

type TableNamespaceConfigurator interface {
	GetTTL() int
	GetLikeString() string
	GetObjectName(string) string
	IsAllowed(string) bool
	Match(string, string, string, string) (*dto.TxnControlCounters, bool)
	RenderTemplate(string) (string, error)
}

var (
	_ TableNamespaceConfigurator = &regexTableNamespaceConfigurator{}
)

type regexTableNamespaceConfigurator struct {
	sqlEngine  sqlengine.SQLEngine
	regex      *regexp.Regexp
	template   *template.Template
	likeString string
	ttl        int
}

func (stc *regexTableNamespaceConfigurator) IsAllowed(tableString string) bool {
	return stc.isAllowed(tableString)
}

func (stc *regexTableNamespaceConfigurator) GetTTL() int {
	return stc.ttl
}

func (stc *regexTableNamespaceConfigurator) GetLikeString() string {
	return stc.getLikeString()
}

func (stc *regexTableNamespaceConfigurator) getLikeString() string {
	return stc.likeString
}

func (stc *regexTableNamespaceConfigurator) isAllowed(tableString string) bool {
	return stc.regex.MatchString(tableString)
}

func (stc *regexTableNamespaceConfigurator) Match(tableString string, requestEncoding string, lastModifiedColName string, requestEncodingColName string) (*dto.TxnControlCounters, bool) {
	isAllowed := stc.isAllowed(tableString)
	if !isAllowed {
		return nil, false
	}
	actualTableName, err := stc.renderTemplate(tableString)
	if err != nil {
		return nil, false
	}
	isPresent := stc.sqlEngine.IsTablePresent(actualTableName, requestEncoding, requestEncodingColName)
	if !isPresent {
		return nil, false
	}
	oldestUpdate, tcc := stc.sqlEngine.TableOldestUpdateUTC(actualTableName, requestEncoding, lastModifiedColName, requestEncodingColName)
	diff := time.Since(oldestUpdate)
	ds := diff.Seconds()
	if stc.ttl > -1 && int(ds) > stc.ttl {
		return nil, false
	}
	return tcc, true
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
