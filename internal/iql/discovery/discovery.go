package discovery

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	"encoding/json"
	"io/ioutil"
	"net/http"

	"infraql/internal/iql/cache"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/googlediscovery"
	"infraql/internal/iql/netutils"
	"infraql/internal/iql/sqlengine"

	"infraql/internal/pkg/openapistackql"

	log "github.com/sirupsen/logrus"
)

const (
	ambiguousServiceErrorMessage string = "More than one service exists with this name, please use the id in the object name, or unset the --usenonpreferredapis flag"
)

type IDiscoveryStore interface {
	ProcessProviderDiscoveryDoc(string, string, dto.RuntimeCtx, string, func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Provider, error), cache.IMarshaller) (*openapistackql.Provider, error)
	ProcessServiceDiscoveryDoc(string, *openapistackql.ProviderService, string, dto.RuntimeCtx, string, func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Service, error), cache.IMarshaller) (*openapistackql.Service, error)
}

type TTLDiscoveryStore struct {
	ttlCache  cache.IKeyValCache
	sqlengine sqlengine.SQLEngine
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
	cacheDir           string
	runtimeCtx         *dto.RuntimeCtx
	rootDocParser      func(bytes []byte, dbEngine sqlengine.SQLEngine, alias string) (*openapistackql.Provider, error)
	serviceDocParser   func(bytes []byte, dbEngine sqlengine.SQLEngine, alias string) (*openapistackql.Service, error)
	rootMarshaller     cache.IMarshaller
	serviceMarshaller  cache.IMarshaller
}

func NewBasicDiscoveryAdapter(
	alias string,
	apiDiscoveryDocUrl string,
	discoveryStore IDiscoveryStore,
	cacheDir string,
	runtimeCtx *dto.RuntimeCtx,
	rootDocParser func(bytes []byte, dbEngine sqlengine.SQLEngine, alias string) (*openapistackql.Provider, error),
	serviceDocParser func(bytes []byte, dbEngine sqlengine.SQLEngine, alias string) (*openapistackql.Service, error),
	rootMarshaller cache.IMarshaller,
	serviceMarshaller cache.IMarshaller,
) IDiscoveryAdapter {
	return &BasicDiscoveryAdapter{
		alias:              alias,
		apiDiscoveryDocUrl: apiDiscoveryDocUrl,
		discoveryStore:     discoveryStore,
		cacheDir:           cacheDir,
		runtimeCtx:         runtimeCtx,
		rootDocParser:      rootDocParser,
		serviceDocParser:   serviceDocParser,
		rootMarshaller:     rootMarshaller,
		serviceMarshaller:  serviceMarshaller,
	}
}

func (adp *BasicDiscoveryAdapter) getServiceDiscoveryDoc(providerKey, serviceKey string, runtimeCtx dto.RuntimeCtx) (*openapistackql.Service, error) {
	component, err := adp.GetServiceHandle(providerKey, serviceKey)
	if component == nil || err != nil {
		return nil, err
	}
	return adp.discoveryStore.ProcessServiceDiscoveryDoc(providerKey, component, adp.cacheDir, runtimeCtx, fmt.Sprintf("%s.%s", adp.alias, serviceKey), adp.serviceDocParser, adp.serviceMarshaller)
}

func (adp *BasicDiscoveryAdapter) GetProvider(providerKey string) (*openapistackql.Provider, error) {
	return adp.discoveryStore.ProcessProviderDiscoveryDoc(adp.apiDiscoveryDocUrl, adp.cacheDir, *adp.runtimeCtx, adp.alias, adp.rootDocParser, adp.rootMarshaller)
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
	serviceIdString := googlediscovery.TranslateServiceKeyIqlToGoogle(serviceKey)
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
	disDoc, err := adp.discoveryStore.ProcessServiceDiscoveryDoc(providerKey, component, adp.cacheDir, *adp.runtimeCtx, fmt.Sprintf("%s.%s", adp.alias, serviceKey), adp.serviceDocParser, adp.serviceMarshaller)
	if err != nil {
		return nil, err
	}
	return disDoc.GetResources()
}

func NewTTLDiscoveryStore(dbEngine sqlengine.SQLEngine, runtimeCtx dto.RuntimeCtx, cacheName string, size int, ttl int, marshaller cache.IMarshaller, sqlengine sqlengine.SQLEngine, alias string) IDiscoveryStore {
	return &TTLDiscoveryStore{
		ttlCache:  cache.NewTTLMap(dbEngine, runtimeCtx, cacheName, size, ttl, marshaller),
		sqlengine: sqlengine,
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

func defaultParser(bytes []byte, dbEngine sqlengine.SQLEngine, alias string) (interface{}, error) {
	var result map[string]interface{}
	jsonErr := json.Unmarshal(bytes, &result)
	return result, jsonErr
}

func parseServiceDiscoveryDoc(bodyBytes []byte, dbEngine sqlengine.SQLEngine, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Service, error)) (*openapistackql.Service, error) {
	result, jsonErr := parser(bodyBytes, dbEngine, alias)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return result, nil
}

func parseProviderDiscoveryDoc(bodyBytes []byte, dbEngine sqlengine.SQLEngine, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Provider, error)) (*openapistackql.Provider, error) {
	result, jsonErr := parser(bodyBytes, dbEngine, alias)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return result, nil
}

func processProviderDiscoveryDoc(url string, cacheDir string, fileMode os.FileMode, runtimeCtx dto.RuntimeCtx, dbEngine sqlengine.SQLEngine, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Provider, error)) (*openapistackql.Provider, error) {
	isMatch, _ := regexp.MatchString(`google`, url)
	if isMatch {
		return openapistackql.LoadProviderByName("google")
	}
	body, err := DownloadDiscoveryDoc(url, runtimeCtx)
	if err != nil {
		return nil, err
	}
	if body == nil {
		return nil, fmt.Errorf("error downloading provider discovery document.  Hint: check network settings, proxy config.")
	}
	defer body.Close()
	bodyBytes, readErr := ioutil.ReadAll(body)
	if readErr != nil {
		return nil, readErr
	}

	// TODO: convert to openapistackql
	return parseProviderDiscoveryDoc(bodyBytes, dbEngine, alias, parser)
}

// func is() bool {

// }

func processServiceDiscoveryDoc(url string, cacheDir string, fileMode os.FileMode, runtimeCtx dto.RuntimeCtx, dbEngine sqlengine.SQLEngine, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Service, error)) (*openapistackql.Service, error) {
	pathComponents := strings.Split(url, "/")
	if len(pathComponents) > 1 {
		provDir := pathComponents[0]
		svcDir := pathComponents[1]
		switch provDir {
		case "googleapis.com", "google":
			pr, err := openapistackql.LoadProviderByName("google")
			if err != nil {
				return nil, err
			}
			return pr.GetService(svcDir)
		}
	}
	body, err := DownloadDiscoveryDoc(url, runtimeCtx)
	if err != nil {
		return nil, err
	}
	if body == nil {
		return nil, fmt.Errorf("error downloading service discovery document.  Hint: check network settings, proxy config.")
	}
	defer body.Close()
	bodyBytes, readErr := ioutil.ReadAll(body)
	if readErr != nil {
		return nil, readErr
	}

	// TODO: convert to openapistackql
	return parseServiceDiscoveryDoc(bodyBytes, dbEngine, alias, parser)
}

func (store *TTLDiscoveryStore) ProcessProviderDiscoveryDoc(url string, cacheDir string, runtimeCtx dto.RuntimeCtx, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Provider, error), marshaller cache.IMarshaller) (*openapistackql.Provider, error) {
	switch url {
	case "https://www.googleapis.com/discovery/v1/apis":
		return openapistackql.LoadProviderByName("google")
	}

	fileMode := os.FileMode(runtimeCtx.ProviderRootPathMode)
	val := store.ttlCache.Get(url, marshaller)
	var retVal *openapistackql.Provider
	var err error
	switch rv := val.(type) {
	case *openapistackql.Provider:
		return rv, nil
	default:
		log.Infoln(fmt.Sprintf("coud not retrieve discovery doc from cache, type = %T", val))
	}
	if runtimeCtx.WorkOffline {
		retVal, err = processProviderDiscoveryDocFromLocal(url, cacheDir, store.sqlengine, alias, parser)
		if retVal != nil && err == nil {
			log.Infoln("placing discovery doc into cache")
			store.ttlCache.Put(url, retVal, marshaller)
		} else if err != nil {
			log.Infoln(err.Error())
			err = errors.New("Provider information is not available in offline mode, run the command once without the --offline flag, then try again in offline mode")
		}
		return retVal, err
	}
	retVal, err = processProviderDiscoveryDoc(url, cacheDir, fileMode, runtimeCtx, store.sqlengine, alias, parser)
	if err != nil {
		return nil, err
	}
	log.Infoln("placing discovery doc into cache")
	store.ttlCache.Put(url, retVal, marshaller)
	db, err := store.sqlengine.GetDB()
	if err != nil {
		return nil, err
	}
	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}
	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return retVal, err
}

func (store *TTLDiscoveryStore) ProcessServiceDiscoveryDoc(providerKey string, serviceHandle *openapistackql.ProviderService, cacheDir string, runtimeCtx dto.RuntimeCtx, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Service, error), marshaller cache.IMarshaller) (*openapistackql.Service, error) {
	switch providerKey {
	case "googleapis.com", "google":
		pr, err := openapistackql.LoadProviderByName("google")
		if err != nil {
			return nil, err
		}
		svc, err := pr.GetService(serviceHandle.Name)
		if err != nil {
			svc, err = pr.GetService(serviceHandle.ID)
		}
		return svc, err
	}

	fileMode := os.FileMode(runtimeCtx.ProviderRootPathMode)
	val := store.ttlCache.Get(serviceHandle.ServiceRef.Ref, marshaller)
	var retVal *openapistackql.Service
	var err error
	switch rv := val.(type) {
	case *openapistackql.Service:
		return rv, nil
	default:
		log.Infoln(fmt.Sprintf("coud not retrieve discovery doc from cache, type = %T", val))
	}
	if runtimeCtx.WorkOffline {
		retVal, err = processServiceDiscoveryDocFromLocal(serviceHandle.ServiceRef.Ref, cacheDir, store.sqlengine, alias, parser)
		if retVal != nil && err == nil {
			log.Infoln("placing discovery doc into cache")
			store.ttlCache.Put(serviceHandle.ServiceRef.Ref, retVal, marshaller)
		} else if err != nil {
			log.Infoln(err.Error())
			err = errors.New("Provider information is not available in offline mode, run the command once without the --offline flag, then try again in offline mode")
		}
		return retVal, err
	}
	retVal, err = processServiceDiscoveryDoc(serviceHandle.ServiceRef.Ref, cacheDir, fileMode, runtimeCtx, store.sqlengine, alias, parser)
	if err != nil {
		return nil, err
	}
	log.Infoln("placing discovery doc into cache")
	store.ttlCache.Put(serviceHandle.ServiceRef.Ref, retVal, marshaller)
	db, err := store.sqlengine.GetDB()
	if err != nil {
		return nil, err
	}
	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}
	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return retVal, err
}

func processServiceDiscoveryDocFromLocal(url string, cacheDir string, dbEngine sqlengine.SQLEngine, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Service, error)) (*openapistackql.Service, error) {
	_, fileName := path.Split(url)
	fullPath := path.Join(cacheDir, fileName)
	bodyBytes, readErr := ioutil.ReadFile(fullPath)
	if readErr != nil {
		log.Infoln(fmt.Sprintf(`cannot process discovery doc with url = "%s", cacheDir = "%s", fullPath = "%s"`, url, cacheDir, fullPath))
		return nil, readErr
	}
	return parseServiceDiscoveryDoc(bodyBytes, dbEngine, alias, parser)
}

func processProviderDiscoveryDocFromLocal(url string, cacheDir string, dbEngine sqlengine.SQLEngine, alias string, parser func([]byte, sqlengine.SQLEngine, string) (*openapistackql.Provider, error)) (*openapistackql.Provider, error) {
	_, fileName := path.Split(url)
	fullPath := path.Join(cacheDir, fileName)
	bodyBytes, readErr := ioutil.ReadFile(fullPath)
	if readErr != nil {
		log.Infoln(fmt.Sprintf(`cannot process discovery doc with url = "%s", cacheDir = "%s", fullPath = "%s"`, url, cacheDir, fullPath))
		return nil, readErr
	}
	return parseProviderDiscoveryDoc(bodyBytes, dbEngine, alias, parser)
}
