package tablenamespace

import (
	"regexp"
	"time"
)

var (
	_ TableNamespaceConfiguratorBuilder = &standardTableNamespaceConfiguratorBuilder{}
)

type TableNamespaceConfiguratorBuilder interface {
	Build() (TableNamespaceConfigurator, error)
	WithExpiryTime(expiryTime time.Time) TableNamespaceConfiguratorBuilder
	WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder
}

type standardTableNamespaceConfiguratorBuilder struct {
	regex      *regexp.Regexp
	expiryTime time.Time
}

func newTableNamespaceConfiguratorBuilder() TableNamespaceConfiguratorBuilder {
	return &standardTableNamespaceConfiguratorBuilder{}
}

func (b *standardTableNamespaceConfiguratorBuilder) WithRegexp(regex *regexp.Regexp) TableNamespaceConfiguratorBuilder {
	b.regex = regex
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) WithExpiryTime(expiryTime time.Time) TableNamespaceConfiguratorBuilder {
	b.expiryTime = expiryTime
	return b
}

func (b *standardTableNamespaceConfiguratorBuilder) Build() (TableNamespaceConfigurator, error) {
	return &RegexTableNamespaceConfigurator{
		regex: b.regex,
	}, nil
}
