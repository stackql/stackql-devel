package tablenamespace

import (
	"fmt"
	"regexp"
)

type TableNamespaceConfigurator interface {
	GetTableName(string) string
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

func (stc *RegexTableNamespaceConfigurator) GetTableName(tableString string) string {
	return fmt.Sprintf("%s.%s", stc.prefix, tableString)
}
