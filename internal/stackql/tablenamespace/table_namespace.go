package tablenamespace

import (
	"regexp"
)

type TableNamespaceConfigurator interface {
	GetObjectName(string) string
	Match(string) bool
}

var (
	_ TableNamespaceConfigurator = &RegexTableNamespaceConfigurator{}
)

type RegexTableNamespaceConfigurator struct {
	regex  *regexp.Regexp
	prefix string
}

func (stc *RegexTableNamespaceConfigurator) Match(tableString string) bool {
	return stc.regex.MatchString(tableString)
}

func (stc *RegexTableNamespaceConfigurator) GetObjectName(inputString string) string {
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
