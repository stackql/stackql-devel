package tablenamespace

import (
	"bytes"
	"regexp"
	"text/template"
)

type TableNamespaceConfigurator interface {
	GetObjectName(string) string
	Match(string) bool
	RenderTemplate(string) (string, error)
}

var (
	_ TableNamespaceConfigurator = &RegexTableNamespaceConfigurator{}
)

type RegexTableNamespaceConfigurator struct {
	regex    *regexp.Regexp
	template *template.Template
}

func (stc *RegexTableNamespaceConfigurator) Match(tableString string) bool {
	return stc.regex.MatchString(tableString)
}

func (stc *RegexTableNamespaceConfigurator) RenderTemplate(input string) (string, error) {
	objName := stc.getObjectName(input)
	inputMap := map[string]interface{}{
		"objectName": objName,
	}
	return stc.render(inputMap)
}

func (stc *RegexTableNamespaceConfigurator) render(input map[string]interface{}) (string, error) {
	var tplWr bytes.Buffer
	if err := stc.template.Execute(&tplWr, input); err != nil {
		return "", err
	}
	return tplWr.String(), nil
}

func (stc *RegexTableNamespaceConfigurator) GetObjectName(inputString string) string {
	return stc.getObjectName(inputString)
}

func (stc *RegexTableNamespaceConfigurator) getObjectName(inputString string) string {
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
