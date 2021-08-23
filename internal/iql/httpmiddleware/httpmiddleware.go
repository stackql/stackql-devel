package httpmiddleware

import (
	"fmt"
	"infraql/internal/iql/handler"
	"infraql/internal/iql/httpexec"
	"infraql/internal/iql/provider"
	"net/http"
)

func HttpApiCall(handlerCtx handler.HandlerContext, prov provider.IProvider, requestCtx httpexec.IHttpContext) (*http.Response, error) {
	authCtx, authErr := handlerCtx.GetAuthContext(prov.GetProviderString())
	if authErr != nil {
		return nil, authErr
	}
	httpClient, httpClientErr := prov.Auth(authCtx, authCtx.Type, false)
	if httpClientErr != nil {
		return nil, httpClientErr
	}
	r, err := httpexec.HTTPApiCall(httpClient, requestCtx)
	if handlerCtx.RuntimeContext.HTTPLogEnabled {
		if r != nil {
			handlerCtx.OutErrFile.Write([]byte(fmt.Sprintln(fmt.Sprintf("http response status: %s", r.Status))))
		} else {
			handlerCtx.OutErrFile.Write([]byte(fmt.Sprintln("http response came buck null")))
		}
	}
	if err != nil {
		if handlerCtx.RuntimeContext.HTTPLogEnabled {
			handlerCtx.OutErrFile.Write([]byte(fmt.Sprintln(fmt.Sprintf("http response error: %s", err.Error()))))
		}
		return nil, err
	}
	return r, err
}
