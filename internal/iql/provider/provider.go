package provider

import (
	"fmt"
	"path"

	"net/http"

	"infraql/internal/iql/cache"
	"infraql/internal/iql/config"
	"infraql/internal/iql/constants"
	"infraql/internal/iql/discovery"
	"infraql/internal/iql/docparser"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/httpexec"
	"infraql/internal/iql/methodselect"
	"infraql/internal/iql/sqlengine"

	"infraql/internal/pkg/openapistackql"
)

const (
	ambiguousServiceErrorMessage string = "More than one service exists with this name, please use the id in the object name, or unset the --usenonpreferredapis flag"
	googleProviderName           string = "google"
	SchemaDelimiter              string = docparser.SchemaDelimiter
)

var DummyAuth bool = false

type ProviderParam struct {
	Id     string
	Type   string
	Format string
}

func GetSupportedProviders(extended bool) map[string]map[string]interface{} {
	retVal := make(map[string]map[string]interface{})
	if extended {
		retVal[googleProviderName] = getGoogleMapExtended()
	} else {
		retVal[googleProviderName] = getGoogleMap()
	}
	return retVal
}

type IProvider interface {
	Auth(authCtx *dto.AuthCtx, authTypeRequested string, enforceRevokeFirst bool) (*http.Client, error)

	AuthRevoke(authCtx *dto.AuthCtx) error

	CheckServiceAccountFile(credentialFile string) error

	EnhanceMetadataFilter(string, func(openapistackql.ITable) (openapistackql.ITable, error), map[string]bool) (func(openapistackql.ITable) (openapistackql.ITable, error), error)

	GenerateHTTPRestInstruction(httpContext httpexec.IHttpContext) (httpexec.IHttpContext, error)

	GetCurrentService() string

	GetDefaultKeyForDeleteItems() string

	GetLikeableColumns(string) []string

	GetMethodForAction(serviceName string, resourceName string, iqlAction string, runtimeCtx dto.RuntimeCtx) (*openapistackql.OperationStore, string, error)

	GetMethodSelector() methodselect.IMethodSelector

	GetProviderString() string

	GetProviderServicesRedacted(runtimeCtx dto.RuntimeCtx, extended bool) (map[string]openapistackql.ProviderService, error)

	GetResource(serviceKey string, resourceKey string, runtimeCtx dto.RuntimeCtx) (*openapistackql.Resource, error)

	GetResourcesMap(serviceKey string, runtimeCtx dto.RuntimeCtx) (map[string]*openapistackql.Resource, error)

	GetResourcesRedacted(currentService string, runtimeCtx dto.RuntimeCtx, extended bool) (map[string]*openapistackql.Resource, error)

	GetService(serviceKey string, runtimeCtx dto.RuntimeCtx) (*openapistackql.Service, error)

	GetObjectSchema(serviceName string, resourceName string, schemaName string) (*openapistackql.Schema, error)

	GetSchemaMap(serviceName string, resourceName string) (map[string]*openapistackql.Schema, error)

	GetVersion() string

	InferDescribeMethod(*openapistackql.Resource) (*openapistackql.OperationStore, string, error)

	InferMaxResultsElement(*openapistackql.OperationStore) *dto.HTTPElement

	InferNextPageRequestElement(*openapistackql.OperationStore) *dto.HTTPElement

	InferNextPageResponseElement(*openapistackql.OperationStore) *dto.HTTPElement

	Parameterise(httpContext httpexec.IHttpContext, method *openapistackql.OperationStore, parameters *dto.HttpParameters, requestSchema *openapistackql.Schema) (httpexec.IHttpContext, error)

	SetCurrentService(serviceKey string)

	ShowAuth(authCtx *dto.AuthCtx) (*openapistackql.AuthMetadata, error)

	GetDiscoveryGeneration(sqlengine.SQLEngine) (int, error)
}

func getProviderCacheDir(runtimeCtx dto.RuntimeCtx, providerName string) string {
	return path.Join(runtimeCtx.ProviderRootPath, providerName)
}

func getGoogleProviderCacheDir(runtimeCtx dto.RuntimeCtx) string {
	return getProviderCacheDir(runtimeCtx, googleProviderName)
}

func GetProviderFromRuntimeCtx(runtimeCtx dto.RuntimeCtx, dbEngine sqlengine.SQLEngine) (IProvider, error) {
	providerStr := runtimeCtx.ProviderStr // TODO: support multiple providers
	switch providerStr {
	case config.GetGoogleProviderString():
		return NewGenericProvider(runtimeCtx, providerStr, dbEngine)
	}
	return nil, fmt.Errorf("provider %s not supported", providerStr)
}

func NewGenericProvider(rtCtx dto.RuntimeCtx, providerStr string, dbEngine sqlengine.SQLEngine) (IProvider, error) {
	ttl := rtCtx.CacheTTL
	if rtCtx.WorkOffline {
		ttl = -1
	}
	methSel, err := methodselect.NewMethodSelector(googleProviderName, constants.GoogleV1String)
	if err != nil {
		return nil, err
	}

	da := discovery.NewBasicDiscoveryAdapter(
		rtCtx.ProviderStr, // TODO: allow multiple
		constants.GoogleV1DiscoveryDoc,
		discovery.NewTTLDiscoveryStore(
			dbEngine,
			rtCtx, constants.GoogleV1ProviderCacheName,
			rtCtx.CacheKeyCount, ttl, &cache.RootDiscoveryMarshaller{},
			dbEngine, rtCtx.ProviderStr, // TODO: allow multiple
		),
		getGoogleProviderCacheDir(rtCtx),
		&rtCtx,
		docparser.OpenapiStackQLRootDiscoveryDocParser,
		docparser.OpenapiStackQLServiceDiscoveryDocParser,
		&cache.RootDiscoveryMarshaller{},
		&cache.ServiceDiscoveryMarshaller{},
	)

	p, err := da.GetProvider(rtCtx.ProviderStr)

	if err != nil {
		return nil, err
	}

	gp := &GenericProvider{
		provider:         p,
		runtimeCtx:       rtCtx,
		discoveryAdapter: da,
		apiVersion:       constants.GoogleV1String,
		methodSelector:   methSel,
	}
	return gp, err
}
