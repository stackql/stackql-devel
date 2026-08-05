package provider

import (
	"fmt"
	"net/http"

	"github.com/stackql/any-sdk/pkg/dto"
	"github.com/stackql/any-sdk/public/formulation"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/intrinsic"
	"github.com/stackql/stackql/internal/stackql/methodselect"
	"github.com/stackql/stackql/internal/stackql/parserutil"

	sdk_internal_dto "github.com/stackql/any-sdk/pkg/internaldto"
)

type intrinsicProvider struct {
	runtimeCtx dto.RuntimeCtx
}

func newIntrinsicProvider(runtimeCtx dto.RuntimeCtx) IProvider {
	return &intrinsicProvider{runtimeCtx: runtimeCtx}
}

func (ip *intrinsicProvider) Auth(_ *dto.AuthCtx, _ string, _ bool) (*http.Client, error) {
	return nil, fmt.Errorf("auth is not supported for provider '%s'", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) AuthRevoke(_ *dto.AuthCtx) error {
	return nil
}

func (ip *intrinsicProvider) CheckCredentialFile(_ *dto.AuthCtx) error {
	return nil
}

func (ip *intrinsicProvider) GetDefaultHTTPClient() *http.Client {
	return nil
}

func (ip *intrinsicProvider) EnhanceMetadataFilter(
	_ string,
	metadataFilter func(formulation.ITable) (formulation.ITable, error),
	_ map[string]bool,
) (func(formulation.ITable) (formulation.ITable, error), error) {
	return metadataFilter, nil
}

func (ip *intrinsicProvider) GetCurrentService() string {
	return intrinsic.ServiceName
}

func (ip *intrinsicProvider) GetDefaultKeyForDeleteItems() string {
	return "items"
}

func (ip *intrinsicProvider) GetFirstMethodForAction(
	_ string,
	_ string,
	_ string,
	_ dto.RuntimeCtx,
) (formulation.StandardOperationStore, string, error) {
	return nil, "", fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetLikeableColumns(tableName string) []string {
	switch tableName {
	case "SERVICES":
		return []string{"id", "name", "title"}
	case "RESOURCES":
		return []string{"id", "name", "title"}
	case "METHODS":
		return []string{"id", "name", "sqlVerb"}
	case "PROVIDERS":
		return []string{"name"}
	default:
		return nil
	}
}

func (ip *intrinsicProvider) GetMethodForAction(
	_ string,
	_ string,
	_ string,
	_ parserutil.ColumnKeyedDatastore,
	_ dto.RuntimeCtx,
) (formulation.StandardOperationStore, string, error) {
	return nil, "", fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetMethodSelector() methodselect.IMethodSelector {
	return nil
}

func (ip *intrinsicProvider) GetProvider() (formulation.Provider, error) {
	return nil, fmt.Errorf("provider '%s' has no external registry descriptor", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetProviderString() string {
	return intrinsic.ProviderName
}

func (ip *intrinsicProvider) GetProviderServicesRedacted(
	_ dto.RuntimeCtx,
	_ bool,
) (map[string]formulation.ProviderService, error) {
	return nil, fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetResource(
	_ string,
	_ string,
	_ dto.RuntimeCtx,
) (formulation.Resource, error) {
	return nil, fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetResourcesMap(
	_ string,
	_ dto.RuntimeCtx,
) (map[string]formulation.Resource, error) {
	return nil, fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetResourcesRedacted(
	_ string,
	_ dto.RuntimeCtx,
	_ bool,
) (map[string]formulation.Resource, error) {
	return nil, fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetServiceShard(
	_ string,
	_ string,
	_ dto.RuntimeCtx,
) (formulation.Service, error) {
	return nil, fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetObjectSchema(
	_ string,
	_ string,
	_ string,
) (formulation.Schema, error) {
	return nil, fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) GetVersion() string {
	return intrinsic.ProviderVersion
}

func (ip *intrinsicProvider) InferDescribeMethod(
	_ formulation.Resource,
) (formulation.StandardOperationStore, string, error) {
	return nil, "", fmt.Errorf("provider '%s' is routed through intrinsic plan generation", intrinsic.ProviderName)
}

func (ip *intrinsicProvider) InferMaxResultsElement(
	_ formulation.OperationStore,
) sdk_internal_dto.HTTPElement {
	return nil
}

func (ip *intrinsicProvider) InferNextPageRequestElement(_ internaldto.Heirarchy) sdk_internal_dto.HTTPElement {
	return nil
}

func (ip *intrinsicProvider) InferNextPageResponseElement(_ internaldto.Heirarchy) sdk_internal_dto.HTTPElement {
	return nil
}

func (ip *intrinsicProvider) PersistStaticExternalSQLDataSource(_ dto.RuntimeCtx) error {
	return nil
}

func (ip *intrinsicProvider) SetCurrentService(_ string) {
}

func (ip *intrinsicProvider) ShowAuth(_ *dto.AuthCtx) (*formulation.AuthMetadata, error) {
	return nil, fmt.Errorf("auth metadata not supported for provider '%s'", intrinsic.ProviderName)
}
