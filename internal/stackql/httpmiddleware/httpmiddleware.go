package httpmiddleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/provider"
)

func GetAuthenticatedClient(handlerCtx handler.HandlerContext, prov provider.IProvider) (*http.Client, error) {
	return getAuthenticatedClient(handlerCtx, prov)
}

func getAuthenticatedClient(handlerCtx handler.HandlerContext, prov provider.IProvider) (*http.Client, error) {
	authCtx, authErr := handlerCtx.GetAuthContext(prov.GetProviderString())
	if authErr != nil {
		return nil, authErr
	}
	httpClient, httpClientErr := prov.Auth(authCtx, authCtx.Type, false)
	if httpClientErr != nil {
		return nil, httpClientErr
	}
	return httpClient, nil
}

//nolint:nestif,mnd // acceptable for now
func parseReponseBodyIfErroneous(response *http.Response) (string, error) {
	if response != nil {
		if response.StatusCode >= 300 {
			if response.Body != nil {
				bodyBytes, bErr := io.ReadAll(response.Body)
				if bErr != nil {
					return "", bErr
				}
				bodyStr := string(bodyBytes)
				response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if len(bodyStr) > 0 {
					return fmt.Sprintf("http response status code: %d, response body: %s", response.StatusCode, bodyStr), nil
				}
			}
			return fmt.Sprintf("http response status code: %d, response body is nil", response.StatusCode), nil
		}
	}
	return "", nil
}

//nolint:nestif // acceptable for now
func parseReponseBodyIfPresent(response *http.Response) (string, error) {
	if response != nil {
		if response.Body != nil {
			bodyBytes, bErr := io.ReadAll(response.Body)
			if bErr != nil {
				return "", bErr
			}
			bodyStr := string(bodyBytes)
			response.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyStr) > 0 {
				return fmt.Sprintf("http response status code: %d, response body: %s", response.StatusCode, bodyStr), nil
			}
			return fmt.Sprintf("http response status code: %d, response body is nil", response.StatusCode), nil
		}
	}
	return "nil response", nil
}
