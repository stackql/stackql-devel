package tablenamespace

import (
	"github.com/stackql/stackql/internal/stackql/dto"
)

type TableNamespaceCollection interface {
	GetAnalyticsCacheTableNamespaceConfigurator() TableNamespaceConfigurator
	GetViewsTableNamespaceConfigurator() TableNamespaceConfigurator
}

func NewStandardTableNamespaceCollection(cfg map[string]*dto.NamespaceCfg) (TableNamespaceCollection, error) {
	// nil dereference protect
	if cfg == nil {
		cfg = map[string]*dto.NamespaceCfg{}
	}
	analyticsCfgDirector := getAnalyticsCacheTableNamespaceConfiguratorBuilderDirector(cfg["analytics"])
	viewsCfgDirector := getViewsTableNamespaceConfiguratorBuilderDirector(cfg["views"])
	err := analyticsCfgDirector.Construct()
	if err != nil {
		return nil, err
	}
	err = viewsCfgDirector.Construct()
	if err != nil {
		return nil, err
	}
	rv := &StandardTableNamespaceCollection{
		analyticsCfg: analyticsCfgDirector.GetResult(),
		viewCfg:      viewsCfgDirector.GetResult(),
	}
	return rv, nil
}

type StandardTableNamespaceCollection struct {
	analyticsCfg TableNamespaceConfigurator
	viewCfg      TableNamespaceConfigurator
}

func (col *StandardTableNamespaceCollection) GetAnalyticsCacheTableNamespaceConfigurator() TableNamespaceConfigurator {
	return col.analyticsCfg
}

func (col *StandardTableNamespaceCollection) GetViewsTableNamespaceConfigurator() TableNamespaceConfigurator {
	return col.viewCfg
}
