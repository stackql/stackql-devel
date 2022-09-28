package tablenamespace

import (
	"regexp"
	"time"

	"github.com/stackql/stackql/internal/stackql/dto"
)

var (
	defaultAnalyticsCacheRegexp = regexp.MustCompile(`^stackql_analytics_(?P<objectName>.*)$`)
	defaultViewsRegexp          = regexp.MustCompile(`^stackql_views\.(?P<objectName>.*)$`)
)

type TableNamespaceConfiguratorBuilderDirector interface {
	Construct() error
	GetResult() TableNamespaceConfigurator
}

func getViewsTableNamespaceConfiguratorBuilderDirector(cfg *dto.NamespaceCfg) TableNamespaceConfiguratorBuilderDirector {
	return &viewsTableNamespaceConfiguratorBuilderDirector{
		cfg: cfg,
	}
}

func getAnalyticsCacheTableNamespaceConfiguratorBuilderDirector(cfg *dto.NamespaceCfg) TableNamespaceConfiguratorBuilderDirector {
	return &analyticsCacheTableNamespaceConfiguratorBuilderDirector{
		cfg: cfg,
	}
}

type viewsTableNamespaceConfiguratorBuilderDirector struct {
	cfg               *dto.NamespaceCfg
	viewsConfigurator TableNamespaceConfigurator
}

func (dr *viewsTableNamespaceConfiguratorBuilderDirector) Construct() error {
	viewsRegexp := defaultViewsRegexp
	viewsExpiryTime := time.Now().Add(24 * time.Hour)
	bldr := newTableNamespaceConfiguratorBuilder().WithRegexp(viewsRegexp).WithExpiryTime(viewsExpiryTime)
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
	cfg                   *dto.NamespaceCfg
	analyticsConfigurator TableNamespaceConfigurator
}

func (dr *analyticsCacheTableNamespaceConfiguratorBuilderDirector) Construct() error {
	analytisCacheRegexp := defaultAnalyticsCacheRegexp
	analyticsExpiryTime := time.Now().Add(24 * time.Hour)
	bldr := newTableNamespaceConfiguratorBuilder().WithRegexp(analytisCacheRegexp).WithExpiryTime(analyticsExpiryTime)
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
