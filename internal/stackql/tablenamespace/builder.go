package tablenamespace

import (
	"regexp"
)

var (
	_ TableNamespaceConfiguratorBuilder = &standardTableNamespaceConfiguratorBuilder{}
)

type TableNamespaceConfiguratorBuilder interface {
	Build() (TableNamespaceConfigurator, error)
	WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder
}

type standardTableNamespaceConfiguratorBuilder struct {
	regex *regexp.Regexp
}

func newTableNamespaceConfiguratorBuilder() TableNamespaceConfiguratorBuilder {
	return &standardTableNamespaceConfiguratorBuilder{}
}

func (b *standardTableNamespaceConfiguratorBuilder) WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder {
	b.regex = regex
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) Build() (TableNamespaceConfigurator, error) {
	return &RegexTableNamespaceConfigurator{
		regex: b.regex,
	}, nil
}
