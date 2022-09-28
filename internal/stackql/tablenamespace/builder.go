package tablenamespace

import (
	"regexp"
	"text/template"
)

var (
	_ TableNamespaceConfiguratorBuilder = &standardTableNamespaceConfiguratorBuilder{}
)

type TableNamespaceConfiguratorBuilder interface {
	Build() (TableNamespaceConfigurator, error)
	WithTTL(ttl int) TableNamespaceConfiguratorBuilder
	WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder
	WithTemplate(regex *template.Template) TableNamespaceConfiguratorBuilder
}

type standardTableNamespaceConfiguratorBuilder struct {
	regex *regexp.Regexp
	tmpl  *template.Template
	ttl   int
}

func newTableNamespaceConfiguratorBuilder() TableNamespaceConfiguratorBuilder {
	return &standardTableNamespaceConfiguratorBuilder{}
}

func (b *standardTableNamespaceConfiguratorBuilder) WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder {
	b.regex = regex
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) WithTemplate(tmpl *template.Template) TableNamespaceConfiguratorBuilder {
	b.tmpl = tmpl
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) WithTTL(ttl int) TableNamespaceConfiguratorBuilder {
	b.ttl = ttl
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) Build() (TableNamespaceConfigurator, error) {
	return &RegexTableNamespaceConfigurator{
		regex: b.regex,
	}, nil
}
