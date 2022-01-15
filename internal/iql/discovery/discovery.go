package discovery

import (
	"fmt"
	"io"

	"net/http"

	"github.com/stackql/stackql/internal/iql/docparser"
	"github.com/stackql/stackql/internal/iql/dto"
	"github.com/stackql/stackql/internal/iql/netutils"
	"github.com/stackql/stackql/internal/iql/sqlengine"

	"github.com/stackql/go-openapistackql/openapistackql"
)

type IDiscoveryStore interface {
	ProcessProviderDiscoveryDoc(string, string) (*openapistackql.Provider, error)
	ProcessServiceDiscoveryDoc(string, *openapistackql.ProviderService, string) (*openapistackql.Service, error)
}

type TTLDiscoveryStore struct {
	sqlengine  sqlengine.SQLEngine
	runtimeCtx dto.RuntimeCtx
}

type IDiscoveryAdapter interface {
	GetResourcesMap(providerKey, serviceKey string) (map[string]*openapistackql.Resource, error)
	GetSchemaMap(providerName, serviceName string, resourceName string) (map[string]*openapistackql.Schema, error)
	GetService(providerKey, serviceKey string) (*openapistackql.Service, error)
	GetServiceHandlesMap(providerKey string) (map[string]openapistackql.ProviderService, error)
	GetServiceHandle(providerKey, serviceKey string) (*openapistackql.ProviderService, error)
	GetProvider(providerKey string) (*openapistackql.Provider, error)
}

type BasicDiscoveryAdapter struct {
	alias              string
	apiDiscoveryDocUrl string
	discoveryStore     IDiscoveryStore
	runtimeCtx         *dto.RuntimeCtx
}

func NewBasicDiscoveryAdapter(
	alias string,
	apiDiscoveryDocUrl string,
	discoveryStore IDiscoveryStore,
	runtimeCtx *dto.RuntimeCtx,
) IDiscoveryAdapter {
	return &BasicDiscoveryAdapter{
		alias:              alias,
		apiDiscoveryDocUrl: apiDiscoveryDocUrl,
		discoveryStore:     discoveryStore,
		runtimeCtx:         runtimeCtx,
	}
}

func (adp *BasicDiscoveryAdapter) getServiceDiscoveryDoc(providerKey, serviceKey string, runtimeCtx dto.RuntimeCtx) (*openapistackql.Service, error) {
	component, err := adp.GetServiceHandle(providerKey, serviceKey)
	if component == nil || err != nil {
		return nil, err
	}
	return adp.discoveryStore.ProcessServiceDiscoveryDoc(providerKey, component, fmt.Sprintf("%s.%s", adp.alias, serviceKey))
}

func (adp *BasicDiscoveryAdapter) GetProvider(providerKey string) (*openapistackql.Provider, error) {
	return adp.discoveryStore.ProcessProviderDiscoveryDoc(adp.apiDiscoveryDocUrl, adp.alias)
}

func (adp *BasicDiscoveryAdapter) GetServiceHandlesMap(providerKey string) (map[string]openapistackql.ProviderService, error) {
	disDoc, err := adp.GetProvider(providerKey)
	if err != nil {
		return nil, err
	}
	return disDoc.ProviderServices, err
}

func (adp *BasicDiscoveryAdapter) GetServiceHandle(providerKey, serviceKey string) (*openapistackql.ProviderService, error) {
	ps, err := adp.GetServiceHandlesMap(providerKey)
	if err != nil {
		return nil, err
	}
	rv, ok := ps[serviceKey]
	if !ok {
		return nil, fmt.Errorf("could not find providerService = '%s'", serviceKey)
	}
	return &rv, nil
}

func (adp *BasicDiscoveryAdapter) GetSchemaMap(providerName string, serviceName string, resourceName string) (map[string]*openapistackql.Schema, error) {
	svcDiscDocMap, err := adp.getServiceDiscoveryDoc(providerName, serviceName, *adp.runtimeCtx)
	if err != nil {
		return nil, err
	}
	return svcDiscDocMap.GetSchemas()
}

func (adp *BasicDiscoveryAdapter) GetService(providerKey, serviceKey string) (*openapistackql.Service, error) {
	serviceIdString := docparser.TranslateServiceKeyIqlToGenericProvider(serviceKey)
	sh, err := adp.GetServiceHandle(providerKey, serviceIdString)
	if err != nil {
		return nil, err
	}
	return sh.GetService()
}

func (adp *BasicDiscoveryAdapter) GetResourcesMap(providerKey, serviceKey string) (map[string]*openapistackql.Resource, error) {
	component, err := adp.GetServiceHandle(providerKey, serviceKey)
	if component == nil || err != nil {
		return nil, err
	}
	disDoc, err := adp.discoveryStore.ProcessServiceDiscoveryDoc(providerKey, component, fmt.Sprintf("%s.%s", adp.alias, serviceKey))
	if err != nil {
		return nil, err
	}
	return disDoc.GetResources()
}

func NewTTLDiscoveryStore(sqlengine sqlengine.SQLEngine, runtimeCtx dto.RuntimeCtx) IDiscoveryStore {
	return &TTLDiscoveryStore{
		sqlengine:  sqlengine,
		runtimeCtx: runtimeCtx,
	}
}

func DownloadDiscoveryDoc(url string, runtimeCtx dto.RuntimeCtx) (io.ReadCloser, error) {
	httpClient := netutils.GetHttpClient(runtimeCtx, nil)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("discovery doc download for '%s' failed with code = %d", url, res.StatusCode)
	}
	return res.Body, nil
}

func (store *TTLDiscoveryStore) ProcessProviderDiscoveryDoc(url string, alias string) (*openapistackql.Provider, error) {
	switch url {
	case "https://www.googleapis.com/discovery/v1/apis":
		return openapistackql.LoadProviderByName("google")
	case "okta":
		return openapistackql.LoadProviderByName("okta")
	}
	return nil, fmt.Errorf("cannot process provider discovery doc url = '%s'", url)
}

func (store *TTLDiscoveryStore) ProcessServiceDiscoveryDoc(providerKey string, serviceHandle *openapistackql.ProviderService, alias string) (*openapistackql.Service, error) {
	// k := fmt.Sprintf("%s.%s", providerKey, serviceHandle.Name)
	switch providerKey {
	case "googleapis.com", "google":
		k := fmt.Sprintf("%s.%s", "google", serviceHandle.Name)
		b, err := store.sqlengine.CacheStoreGet(k)
		if b != nil && err == nil {
			return openapistackql.LoadServiceDocFromBytes(b)
		}
		pr, err := openapistackql.LoadProviderByName("google")
		if err != nil {
			return nil, err
		}
		svc, err := pr.GetService(serviceHandle.Name)
		if err != nil {
			svc, err = pr.GetService(serviceHandle.ID)
		}
		bt, err := svc.ToYaml()
		if err != nil {
			return nil, err
		}
		err = store.sqlengine.CacheStorePut(k, bt, "", 0)
		if err != nil {
			return nil, err
		}
		err = docparser.OpenapiStackQLServiceDiscoveryDocPersistor(pr, svc, store.sqlengine, pr.Name)
		return svc, err
	default:
		k := fmt.Sprintf("%s.%s", providerKey, serviceHandle.Name)
		b, err := store.sqlengine.CacheStoreGet(k)
		if b != nil && err == nil {
			return openapistackql.LoadServiceDocFromBytes(b)
		}
		pr, err := openapistackql.LoadProviderByName(providerKey)
		if err != nil {
			return nil, err
		}
		svc, err := pr.GetService(serviceHandle.Name)
		if err != nil {
			svc, err = pr.GetService(serviceHandle.ID)
		}
		bt, err := svc.ToYaml()
		if err != nil {
			return nil, err
		}
		err = store.sqlengine.CacheStorePut(k, bt, "", 0)
		if err != nil {
			return nil, err
		}
		err = docparser.OpenapiStackQLServiceDiscoveryDocPersistor(pr, svc, store.sqlengine, pr.Name)
		return svc, err
	}
}

// func (store *TTLDiscoveryStore) GetService(providerKey string, serviceHandle *openapistackql.ProviderService, alias string) (*openapistackql.Service, error) {

// }
