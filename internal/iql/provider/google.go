package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"infraql/internal/iql/constants"
	"infraql/internal/iql/discovery"
	"infraql/internal/iql/dto"
	sdk "infraql/internal/iql/google_sdk"
	"infraql/internal/iql/googlediscovery"
	"infraql/internal/iql/httpexec"
	"infraql/internal/iql/methodselect"
	"infraql/internal/iql/netutils"
	"infraql/internal/iql/relational"
	"infraql/internal/iql/sqlengine"

	"infraql/internal/pkg/openapistackql"
	"infraql/internal/pkg/sqltypeutil"

	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	log "github.com/sirupsen/logrus"
)

var (
	storageObjectsRegex *regexp.Regexp = regexp.MustCompile(`^storage\.objects\..*$`)
)

type googleServiceAccount struct {
	Email      string `json:"client_email"`
	PrivateKey string `json:"private_key"`
}

func getGoogleMap() map[string]interface{} {
	googleMap := map[string]interface{}{
		"name": googleProviderName,
	}
	return googleMap
}

func getGoogleMapExtended() map[string]interface{} {
	return getGoogleMap()
}

type GoogleProvider struct {
	runtimeCtx       dto.RuntimeCtx
	currentService   string
	discoveryAdapter discovery.IDiscoveryAdapter
	apiVersion       string
	methodSelector   methodselect.IMethodSelector
}

func (gp *GoogleProvider) getDefaultKeyForSelectItems(sc *openapistackql.Schema) string {
	return "items"
}

func (gp *GoogleProvider) GetDiscoveryGeneration(dbEngine sqlengine.SQLEngine) (int, error) {
	return dbEngine.GetCurrentDiscoveryGenerationId(gp.GetProviderString())
}

func (gp *GoogleProvider) GetDefaultKeyForDeleteItems() string {
	return "items"
}

func (gp *GoogleProvider) GetMethodSelector() methodselect.IMethodSelector {
	return gp.methodSelector
}

func (gp *GoogleProvider) GetVersion() string {
	return gp.apiVersion
}

func (gp *GoogleProvider) GetService(serviceKey string, runtimeCtx dto.RuntimeCtx) (*openapistackql.Service, error) {
	return gp.discoveryAdapter.GetService("google", serviceKey)
}

func (gp *GoogleProvider) inferAuthType(authCtx dto.AuthCtx, authTypeRequested string) string {
	switch strings.ToLower(authTypeRequested) {
	case dto.AuthServiceAccountStr:
		return dto.AuthServiceAccountStr
	case dto.AuthInteractiveStr:
		return dto.AuthInteractiveStr
	}
	if authCtx.KeyFilePath != "" {
		return dto.AuthServiceAccountStr
	}
	return dto.AuthInteractiveStr
}

func (gp *GoogleProvider) Auth(authCtx *dto.AuthCtx, authTypeRequested string, enforceRevokeFirst bool) (*http.Client, error) {
	switch gp.inferAuthType(*authCtx, authTypeRequested) {
	case dto.AuthServiceAccountStr:
		return gp.keyFileAuth(authCtx)
	case dto.AuthInteractiveStr:
		return gp.oAuth(authCtx, enforceRevokeFirst)
	}
	return nil, fmt.Errorf("Could not infer auth type")
}

func (gp *GoogleProvider) AuthRevoke(authCtx *dto.AuthCtx) error {
	switch strings.ToLower(authCtx.Type) {
	case dto.AuthServiceAccountStr:
		return errors.New(constants.ServiceAccountRevokeErrStr)
	case dto.AuthInteractiveStr:
		err := sdk.RevokeGoogleAuth()
		if err == nil {
			deactivateAuth(authCtx)
		}
		return err
	}
	return fmt.Errorf(`Auth revoke for Google Failed; improper auth method: "%s" speciied`, authCtx.Type)
}

func (gp *GoogleProvider) GetMethodForAction(serviceName string, resourceName string, iqlAction string, runtimeCtx dto.RuntimeCtx) (*openapistackql.OperationStore, string, error) {
	rsc, err := gp.GetResource(serviceName, resourceName, runtimeCtx)
	if err != nil {
		return nil, "", err
	}
	return gp.methodSelector.GetMethodForAction(rsc, iqlAction)
}

func (gp *GoogleProvider) InferDescribeMethod(rsc *openapistackql.Resource) (*openapistackql.OperationStore, string, error) {
	if rsc == nil {
		return nil, "", fmt.Errorf("cannot infer describe method from nil resource")
	}
	var method *openapistackql.OperationStore
	m, methodErr := rsc.FindMethod("list")
	if methodErr == nil && m != nil {
		return m, "list", nil
	}
	m, methodErr = rsc.FindMethod("aggregatedList")
	if methodErr == nil && m != nil {
		return m, "aggregatedList", nil
	}
	m, methodErr = rsc.FindMethod("get")
	if methodErr == nil && m != nil {
		return m, "get", nil
	}
	var ms []string
	for _, v := range rsc.Methods {
		vp := &v
		ms = append(ms, v.GetName())
		if strings.HasPrefix(v.GetName(), "get") {
			method = vp
			return method, v.GetName(), nil
		}
	}
	for _, v := range rsc.Methods {
		vp := &v
		if strings.HasPrefix(v.GetName(), "list") {
			method = vp
			return method, v.GetName(), nil
		}
	}
	return nil, "", fmt.Errorf("SELECT not supported for this resource, use SHOW METHODS to view available operations for the resource and then invoke a supported method using the EXEC command")
}

func (gp *GoogleProvider) retrieveSchemaMap(serviceName string, resourceName string) (map[string]*openapistackql.Schema, error) {
	return gp.discoveryAdapter.GetSchemaMap("google", serviceName, resourceName)
}

func (gp *GoogleProvider) GetSchemaMap(serviceName string, resourceName string) (map[string]*openapistackql.Schema, error) {
	return gp.discoveryAdapter.GetSchemaMap("google", serviceName, resourceName)
}

func (gp *GoogleProvider) GetObjectSchema(serviceName string, resourceName string, schemaName string) (*openapistackql.Schema, error) {
	sm, err := gp.retrieveSchemaMap(serviceName, resourceName)
	if err != nil {
		return nil, err
	}
	s := sm[schemaName]
	return s, nil
}

type transport struct {
	token               []byte
	underlyingTransport http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", string(t.token)),
	)
	return t.underlyingTransport.RoundTrip(req)
}

func activateAuth(authCtx *dto.AuthCtx, principal string, authType string) {
	authCtx.Active = true
	authCtx.Type = authType
	if principal != "" {
		authCtx.ID = principal
	}
}

func deactivateAuth(authCtx *dto.AuthCtx) {
	authCtx.Active = false
}

func (gp *GoogleProvider) ShowAuth(authCtx *dto.AuthCtx) (*openapistackql.AuthMetadata, error) {
	var err error
	var retVal *openapistackql.AuthMetadata
	var authObj openapistackql.AuthMetadata
	switch gp.inferAuthType(*authCtx, authCtx.Type) {
	case dto.AuthServiceAccountStr:
		var sa googleServiceAccount
		sa, err = parseServiceAccountFile(authCtx.KeyFilePath)
		if err == nil {
			authObj = openapistackql.AuthMetadata{
				Principal: sa.Email,
				Type:      strings.ToUpper(dto.AuthServiceAccountStr),
				Source:    authCtx.KeyFilePath,
			}
			retVal = &authObj
			activateAuth(authCtx, sa.Email, dto.AuthServiceAccountStr)
		}
	case dto.AuthInteractiveStr:
		principal, sdkErr := sdk.GetCurrentAuthUser()
		if sdkErr == nil {
			principalStr := string(principal)
			if principalStr != "" {
				authObj = openapistackql.AuthMetadata{
					Principal: principalStr,
					Type:      strings.ToUpper(dto.AuthInteractiveStr),
					Source:    "OAuth",
				}
				retVal = &authObj
				activateAuth(authCtx, principalStr, dto.AuthInteractiveStr)
			} else {
				err = errors.New(constants.NotAuthenticatedShowStr)
			}
		} else {
			log.Infoln(sdkErr)
			err = errors.New(constants.NotAuthenticatedShowStr)
		}
	default:
		err = errors.New(constants.NotAuthenticatedShowStr)
	}
	return retVal, err
}

func (gp *GoogleProvider) oAuth(authCtx *dto.AuthCtx, enforceRevokeFirst bool) (*http.Client, error) {
	var err error
	var tokenBytes []byte
	tokenBytes, err = sdk.GetAccessToken()
	if enforceRevokeFirst && authCtx.Type == dto.AuthInteractiveStr && err == nil {
		return nil, fmt.Errorf(constants.OAuthInteractiveAuthErrStr)
	}
	if err != nil {
		err = sdk.OAuthToGoogle()
		if err == nil {
			tokenBytes, err = sdk.GetAccessToken()
		}
	}
	if err != nil {
		return nil, err
	}
	activateAuth(authCtx, "", dto.AuthInteractiveStr)
	client := netutils.GetHttpClient(gp.runtimeCtx, nil)
	client.Transport = &transport{
		token:               tokenBytes,
		underlyingTransport: client.Transport,
	}
	return client, nil
}

func (gp *GoogleProvider) keyFileAuth(authCtx *dto.AuthCtx) (*http.Client, error) {
	scopes := authCtx.Scopes
	if scopes == nil {
		scopes = []string{
			"https://www.googleapis.com/auth/cloud-platform",
		}
	}
	return serviceAccount(authCtx, scopes, gp.runtimeCtx)
}

func (gp *GoogleProvider) getServiceType(service *openapistackql.Service) string {
	specialServiceNamesMap := map[string]bool{
		"storage": true,
		"compute": true,
		"dns":     true,
		"sql":     true,
	}
	nameIsSpecial, ok := specialServiceNamesMap[service.GetName()]
	cloudRegex := regexp.MustCompile(`(^https://.*cloud\.google\.com|^https://firebase\.google\.com)`)
	if service.IsPreferred() && (cloudRegex.MatchString(service.Info.Contact.URL) || (ok && nameIsSpecial)) {
		return "cloud"
	}
	return "developer"
}

func (gp *GoogleProvider) GetLikeableColumns(tableName string) []string {
	var retVal []string
	switch tableName {
	case "SERVICES":
		return []string{
			"id",
			"name",
		}
	case "RESOURCES":
		return []string{
			"id",
			"name",
		}
	case "METHODS":
		return []string{
			"id",
			"name",
		}
	case "PROVIDERS":
		return []string{
			"name",
		}
	}
	return retVal
}

func (gp *GoogleProvider) EnhanceMetadataFilter(metadataType string, metadataFilter func(openapistackql.ITable) (openapistackql.ITable, error), colsVisited map[string]bool) (func(openapistackql.ITable) (openapistackql.ITable, error), error) {
	typeVisited, typeOk := colsVisited["type"]
	preferredVisited, preferredOk := colsVisited["preferred"]
	sqlTrue, sqlTrueErr := sqltypeutil.InterfaceToSQLType(true)
	sqlCloudStr, sqlCloudStrErr := sqltypeutil.InterfaceToSQLType("cloud")
	equalsOperator, operatorErr := relational.GetOperatorPredicate("=")
	if sqlTrueErr != nil || sqlCloudStrErr != nil || operatorErr != nil {
		return nil, fmt.Errorf("typing and operator system broken!!!")
	}
	switch metadataType {
	case "service":
		if typeOk && typeVisited && preferredOk && preferredVisited {
			return metadataFilter, nil
		}
		if typeOk && typeVisited {
			return relational.AndTableFilters(
				metadataFilter,
				relational.ConstructTablePredicateFilter("preferred", sqlTrue, equalsOperator),
			), nil
		}
		if preferredOk && preferredVisited {
			return relational.AndTableFilters(
				metadataFilter,
				relational.ConstructTablePredicateFilter("type", sqlCloudStr, equalsOperator),
			), nil
		}
		return relational.AndTableFilters(
			relational.AndTableFilters(
				metadataFilter,
				relational.ConstructTablePredicateFilter("cloud", sqlCloudStr, equalsOperator),
			),
			relational.ConstructTablePredicateFilter("preferred", sqlTrue, equalsOperator),
		), nil
	}
	return metadataFilter, nil
}

func (gp *GoogleProvider) getProviderServices() (map[string]openapistackql.ProviderService, error) {
	retVal := make(map[string]openapistackql.ProviderService)
	disDoc, err := gp.discoveryAdapter.GetServiceHandlesMap("google")
	if err != nil {
		return nil, err
	}
	for k, item := range disDoc {
		retVal[googlediscovery.TranslateServiceKeyGoogleToIql(k)] = item
	}
	return retVal, nil
}

func (gp *GoogleProvider) GetProviderServicesRedacted(runtimeCtx dto.RuntimeCtx, extended bool) (map[string]openapistackql.ProviderService, error) {
	return gp.getProviderServices()
}

func (gp *GoogleProvider) GetResourcesRedacted(currentService string, runtimeCtx dto.RuntimeCtx, extended bool) (map[string]*openapistackql.Resource, error) {
	svcDiscDocMap, err := gp.discoveryAdapter.GetResourcesMap("google", currentService)
	return svcDiscDocMap, err
}

func parseServiceAccountFile(credentialFile string) (googleServiceAccount, error) {
	b, err := ioutil.ReadFile(credentialFile)
	var c googleServiceAccount
	if err != nil {
		return c, errors.New(constants.ServiceAccountPathErrStr)
	}
	return c, json.Unmarshal(b, &c)
}

func (gp *GoogleProvider) CheckServiceAccountFile(credentialFile string) error {
	_, err := parseServiceAccountFile(credentialFile)
	return err
}

func serviceAccount(authCtx *dto.AuthCtx, scopes []string, runtimeCtx dto.RuntimeCtx) (*http.Client, error) {
	credentialFile := authCtx.KeyFilePath
	b, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		return nil, errors.New(constants.ServiceAccountPathErrStr)
	}
	config, errToken := google.JWTConfigFromJSON(b, scopes...)
	if errToken != nil {
		return nil, errToken
	}
	activateAuth(authCtx, "", dto.AuthServiceAccountStr)
	httpClient := netutils.GetHttpClient(runtimeCtx, http.DefaultClient)
	if DummyAuth {
		// return httpClient, nil
	}
	return config.Client(context.WithValue(oauth2.NoContext, oauth2.HTTPClient, httpClient)), nil
}

func (gp *GoogleProvider) GenerateHTTPRestInstruction(httpContext httpexec.IHttpContext) (httpexec.IHttpContext, error) {
	return httpContext, nil
}

func (gp *GoogleProvider) escapeUrlParameter(k string, v string, method *openapistackql.OperationStore) string {
	if storageObjectsRegex.MatchString(method.GetName()) {
		return url.QueryEscape(v)
	}
	return v
}

func (gp *GoogleProvider) Parameterise(httpContext httpexec.IHttpContext, method *openapistackql.OperationStore, parameters *dto.HttpParameters, requestSchema *openapistackql.Schema) (httpexec.IHttpContext, error) {
	visited := make(map[string]bool)
	args := make([]string, len(parameters.PathParams)*2)
	var sb strings.Builder
	var queryParams []string
	i := 0
	for k, v := range parameters.PathParams {
		if strings.Contains(httpContext.GetTemplateUrl(), "{"+k+"}") {
			args[i] = "{" + k + "}"
			args[i+1] = gp.escapeUrlParameter(k, fmt.Sprint(v), method)
			i += 2
			visited[k] = true
			continue
		}
		if strings.Contains(httpContext.GetTemplateUrl(), "{+"+k+"}") {
			args[i] = "{+" + k + "}"
			args[i+1] = gp.escapeUrlParameter(k, fmt.Sprint(v), method)
			i += 2
			visited[k] = true
			continue
		}
	}
	if len(parameters.QueryParams) > 0 {
		sb.WriteString("?")
	}
	for k, v := range parameters.QueryParams {
		vStr, vOk := v.Val.(string)
		if isVisited, kExists := visited[k]; !kExists || (!isVisited && vOk) {
			queryParams = append(queryParams, k+"="+url.QueryEscape(vStr))
			visited[k] = true
		}
	}
	sb.WriteString(strings.Join(queryParams, "&"))
	httpContext.SetUrl(strings.NewReplacer(args...).Replace(httpContext.GetTemplateUrl()) + sb.String())
	return httpContext, nil
}

func (gp *GoogleProvider) SetCurrentService(serviceKey string) {
	gp.currentService = serviceKey

}

func (gp *GoogleProvider) GetCurrentService() string {
	return gp.currentService
}

func (gp *GoogleProvider) getPathParams(httpContext httpexec.IHttpContext) map[string]bool {
	re := regexp.MustCompile(`\{([^\{\}]+)\}`)
	keys := re.FindAllString(httpContext.GetTemplateUrl(), -1)
	retVal := make(map[string]bool, len(keys))
	for _, k := range keys {
		retVal[strings.Trim(k, "{}")] = true
	}
	return retVal
}

func (gp *GoogleProvider) GetResourcesMap(serviceKey string, runtimeCtx dto.RuntimeCtx) (map[string]*openapistackql.Resource, error) {
	return gp.discoveryAdapter.GetResourcesMap("google", serviceKey)
}

func (gp *GoogleProvider) GetResource(serviceKey string, resourceKey string, runtimeCtx dto.RuntimeCtx) (*openapistackql.Resource, error) {
	rm, err := gp.GetResourcesMap(serviceKey, runtimeCtx)
	retVal, ok := rm[resourceKey]
	if !ok {
		return nil, fmt.Errorf("Could not obtain resource '%s' from service '%s'", resourceKey, serviceKey)
	}
	return retVal, err
}

func (gp *GoogleProvider) GetProviderString() string {
	return googleProviderName
}

func (gp *GoogleProvider) InferMaxResultsElement(*openapistackql.OperationStore) *dto.HTTPElement {
	return &dto.HTTPElement{
		Type: dto.QueryParam,
		Name: "maxResults",
	}
}

func (gp *GoogleProvider) InferNextPageRequestElement(*openapistackql.OperationStore) *dto.HTTPElement {
	return &dto.HTTPElement{
		Type: dto.QueryParam,
		Name: "pageToken",
	}
}

func (gp *GoogleProvider) InferNextPageResponseElement(*openapistackql.OperationStore) *dto.HTTPElement {
	return &dto.HTTPElement{
		Type: dto.BodyAttribute,
		Name: "nextPageToken",
	}
}
