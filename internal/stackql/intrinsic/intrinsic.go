package intrinsic

import (
	"strings"

	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/typing"
	"github.com/stackql/stackql/internal/stackql/util"

	"github.com/stackql/stackql-parser/go/vt/sqlparser"
)

const (
	ProviderName      = "stackql_intrinsic"
	ProviderVersion   = "internal"
	ServiceName       = "audit"
	ResourceName      = "info"
	SelectMethodName  = "select"
	TitleColumn       = "title"
	DescriptionColumn = "descriptiion"
)

func IsProvider(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), ProviderName)
}

func resolveProvider(providerName string, currentProvider string) string {
	if strings.TrimSpace(providerName) != "" {
		return providerName
	}
	return currentProvider
}

func IsInfoTable(tableName sqlparser.TableName, currentProvider string) bool {
	providerName := resolveProvider(tableName.QualifierSecond.GetRawVal(), currentProvider)
	serviceName := tableName.Qualifier.GetRawVal()
	resourceName := tableName.Name.GetRawVal()
	return IsProvider(providerName) && strings.EqualFold(serviceName, ServiceName) && strings.EqualFold(resourceName, ResourceName)
}

func IsShowServices(node *sqlparser.Show, currentProvider string) bool {
	providerName := resolveProvider(node.OnTable.Name.GetRawVal(), currentProvider)
	return strings.EqualFold(strings.ToUpper(node.Type), "SERVICES") && IsProvider(providerName)
}

func IsShowResources(node *sqlparser.Show, currentProvider string) bool {
	providerName := resolveProvider(node.OnTable.Qualifier.GetRawVal(), currentProvider)
	serviceName := node.OnTable.Name.GetRawVal()
	return strings.EqualFold(strings.ToUpper(node.Type), "RESOURCES") &&
		IsProvider(providerName) &&
		strings.EqualFold(serviceName, ServiceName)
}

func IsShowMethods(node *sqlparser.Show, currentProvider string) bool {
	providerName := resolveProvider(node.OnTable.QualifierSecond.GetRawVal(), currentProvider)
	serviceName := node.OnTable.Qualifier.GetRawVal()
	resourceName := node.OnTable.Name.GetRawVal()
	return strings.EqualFold(strings.ToUpper(node.Type), "METHODS") &&
		IsProvider(providerName) &&
		strings.EqualFold(serviceName, ServiceName) &&
		strings.EqualFold(resourceName, ResourceName)
}

func IsDescribeInfo(node *sqlparser.DescribeTable, currentProvider string) bool {
	return IsInfoTable(node.Table, currentProvider)
}

func IsDescribeSelectMethod(node *sqlparser.DescribeMethod, currentProvider string) bool {
	providerName := resolveProvider(node.Provider.GetRawVal(), currentProvider)
	return IsProvider(providerName) &&
		strings.EqualFold(node.Service.GetRawVal(), ServiceName) &&
		strings.EqualFold(node.Resource.GetRawVal(), ResourceName) &&
		strings.EqualFold(node.Method.GetRawVal(), SelectMethodName)
}

func IsDescribeIntrinsicMethod(node *sqlparser.DescribeMethod, currentProvider string) bool {
	providerName := resolveProvider(node.Provider.GetRawVal(), currentProvider)
	return IsProvider(providerName) &&
		strings.EqualFold(node.Service.GetRawVal(), ServiceName) &&
		strings.EqualFold(node.Resource.GetRawVal(), ResourceName)
}

func SelectInfoOutput(typCfg typing.Config) internaldto.ExecutorOutput {
	columnOrder := []string{TitleColumn, DescriptionColumn}
	rows := map[string]map[string]interface{}{
		"000001": {
			TitleColumn:       "stackql intrinsic row 1",
			DescriptionColumn: "dummy intrinsic description row 1",
		},
		"000002": {
			TitleColumn:       "stackql intrinsic row 2",
			DescriptionColumn: "dummy intrinsic description row 2",
		},
	}
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(nil, rows, columnOrder, util.DefaultRowSort, nil, nil, typCfg),
	)
}

func ShowServicesOutput(typCfg typing.Config, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"id", "name", "title"}
	row := map[string]interface{}{
		"id":    ServiceName,
		"name":  ServiceName,
		"title": "Intrinsic Audit",
	}
	if extended {
		columnOrder = append(columnOrder, "description", "version")
		row["description"] = "internal intrinsic service"
		row["version"] = ProviderVersion
	}
	rows := map[string]map[string]interface{}{
		"000001": row,
	}
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(nil, rows, columnOrder, util.DefaultRowSort, nil, nil, typCfg),
	)
}

func ShowResourcesOutput(typCfg typing.Config, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"id", "name", "title"}
	row := map[string]interface{}{
		"id":    ResourceName,
		"name":  ResourceName,
		"title": "Intrinsic Info",
	}
	if extended {
		columnOrder = append(columnOrder, "description")
		row["description"] = "internal intrinsic info resource"
	}
	rows := map[string]map[string]interface{}{
		"000001": row,
	}
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(nil, rows, columnOrder, util.DefaultRowSort, nil, nil, typCfg),
	)
}

func ShowMethodsOutput(typCfg typing.Config, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"id", "name", "sqlVerb"}
	row := map[string]interface{}{
		"id":      SelectMethodName,
		"name":    SelectMethodName,
		"sqlVerb": SelectMethodName,
	}
	if extended {
		columnOrder = append(columnOrder, "description")
		row["description"] = "select-only intrinsic method"
	}
	rows := map[string]map[string]interface{}{
		"000001": row,
	}
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(nil, rows, columnOrder, util.DefaultRowSort, nil, nil, typCfg),
	)
}

func DescribeInfoOutput(typCfg typing.Config, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"name", "type"}
	if extended {
		columnOrder = append(columnOrder, "description")
	}
	rowTitle := map[string]interface{}{
		"name": TitleColumn,
		"type": "string",
	}
	rowDescription := map[string]interface{}{
		"name": DescriptionColumn,
		"type": "string",
	}
	if extended {
		rowTitle["description"] = "intrinsic info title"
		rowDescription["description"] = "intrinsic info description"
	}
	rows := map[string]map[string]interface{}{
		"000001": rowTitle,
		"000002": rowDescription,
	}
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(nil, rows, columnOrder, util.DefaultRowSort, nil, nil, typCfg),
	)
}

func DescribeMethodOutput(typCfg typing.Config, extended bool) internaldto.ExecutorOutput {
	columnOrder := []string{"name", "type", "param_type", "shape"}
	if extended {
		columnOrder = append(columnOrder, "description")
	}
	rowTitle := map[string]interface{}{
		"name":       TitleColumn,
		"type":       "string",
		"param_type": "response",
		"shape":      "string",
	}
	rowDescription := map[string]interface{}{
		"name":       DescriptionColumn,
		"type":       "string",
		"param_type": "response",
		"shape":      "string",
	}
	if extended {
		rowTitle["description"] = "intrinsic info title"
		rowDescription["description"] = "intrinsic info description"
	}
	rows := map[string]map[string]interface{}{
		"000001": rowTitle,
		"000002": rowDescription,
	}
	return util.PrepareResultSet(
		internaldto.NewPrepareResultSetDTO(nil, rows, columnOrder, util.DefaultRowSort, nil, nil, typCfg),
	)
}
