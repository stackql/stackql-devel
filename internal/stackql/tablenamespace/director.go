package tablenamespace

import (
	"regexp"
)

var (
	analyticsCacheRegexp = regexp.MustCompile(`^stackql_analytics_(?P<objectName>.*)$`)
	viewsRegexp          = regexp.MustCompile(`^stackql_views\.(?P<objectName>.*)$`)
)

type TableNamespaceConfiguratorBuilderDirector interface {
	Construct() error
	GetResult() TableNamespaceConfigurator
}

func GetViewsTableNamespaceConfiguratorBuilderDirector() TableNamespaceConfiguratorBuilderDirector {
	return &viewsTableNamespaceConfiguratorBuilderDirector{}
}

func GetAnalyticsCacheTableNamespaceConfiguratorBuilderDirector() TableNamespaceConfiguratorBuilderDirector {
	return &analyticsCacheTableNamespaceConfiguratorBuilderDirector{}
}

type viewsTableNamespaceConfiguratorBuilderDirector struct {
	viewsConfigurator TableNamespaceConfigurator
}

func (dr *viewsTableNamespaceConfiguratorBuilderDirector) Construct() error {
	bldr := newTableNamespaceConfiguratorBuilder().WithRegexp(viewsRegexp)
	configurator, err := bldr.Build()
	if err != nil {
		return err
	}
	dr.viewsConfigurator = configurator
	return nil
}

func (dr *viewsTableNamespaceConfiguratorBuilderDirector) GetResult() TableNamespaceConfigurator {
	return dr.viewsConfigurator
}

type analyticsCacheTableNamespaceConfiguratorBuilderDirector struct {
	analyticsConfigurator TableNamespaceConfigurator
}

func (dr *analyticsCacheTableNamespaceConfiguratorBuilderDirector) Construct() error {
	bldr := newTableNamespaceConfiguratorBuilder().WithRegexp(analyticsCacheRegexp)
	configurator, err := bldr.Build()
	if err != nil {
		return err
	}
	dr.analyticsConfigurator = configurator
	return nil
}

func (dr *analyticsCacheTableNamespaceConfiguratorBuilderDirector) GetResult() TableNamespaceConfigurator {
	return dr.analyticsConfigurator
}
