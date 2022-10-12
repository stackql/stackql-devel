package tablenamespace

import (
	"regexp"
	"text/template"

	"github.com/stackql/stackql/internal/stackql/sqlengine"
)

var (
	_ TableNamespaceConfiguratorBuilder = &standardTableNamespaceConfiguratorBuilder{}
)

type TableNamespaceConfiguratorBuilder interface {
	Build() (TableNamespaceConfigurator, error)
	WithTTL(ttl int) TableNamespaceConfiguratorBuilder
	WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder
	WithSQLEngine(sqlEngine sqlengine.SQLEngine) TableNamespaceConfiguratorBuilder
	WithTemplate(tmpl *template.Template) TableNamespaceConfiguratorBuilder
}

type standardTableNamespaceConfiguratorBuilder struct {
	sqlEngine sqlengine.SQLEngine
	regex     *regexp.Regexp
	tmpl      *template.Template
	ttl       int
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

func (b *standardTableNamespaceConfiguratorBuilder) WithSQLEngine(sqlEngine sqlengine.SQLEngine) TableNamespaceConfiguratorBuilder {
	b.sqlEngine = sqlEngine
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) Build() (TableNamespaceConfigurator, error) {
	return &regexTableNamespaceConfigurator{
		sqlEngine: b.sqlEngine,
		regex:     b.regex,
		template:  b.tmpl,
		ttl:       b.ttl,
	}, nil
}
