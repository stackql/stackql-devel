package netutils

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/stackql/stackql/internal/stackql/dto"
)

func GetRoundTripper(runtimeCtx dto.RuntimeCtx, existingTransport http.RoundTripper) http.RoundTripper {
	return getRoundTripper(runtimeCtx, existingTransport)
}

func getRoundTripper(runtimeCtx dto.RuntimeCtx, existingTransport http.RoundTripper) http.RoundTripper {
	var tr *http.Transport
	var rt http.RoundTripper
	if existingTransport != nil {
		switch exTR := existingTransport.(type) {
		case *http.Transport:
			tr = exTR.Clone()
		default:
			rt = exTR
		}
	} else {
		tr = &http.Transport{}
	}
	host := runtimeCtx.HTTPProxyHost
	if host != "" {
		if runtimeCtx.HTTPProxyPort > 0 {
			host = fmt.Sprintf("%s:%d", runtimeCtx.HTTPProxyHost, runtimeCtx.HTTPProxyPort)
		}
		var usr *url.Userinfo
		if runtimeCtx.HTTPProxyUser != "" {
			usr = url.UserPassword(runtimeCtx.HTTPProxyUser, runtimeCtx.HTTPProxyPassword)
		}
		proxyUrl := &url.URL{
			Host:   host,
			Scheme: runtimeCtx.HTTPProxyScheme,
			User:   usr,
		}
		if tr != nil {
			tr.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	if tr != nil {
		rt = tr
	}
	return rt
}

func GetHttpClient(runtimeCtx dto.RuntimeCtx, existingClient *http.Client) *http.Client {
	return getHttpClient(runtimeCtx, existingClient)
}

func getHttpClient(runtimeCtx dto.RuntimeCtx, existingClient *http.Client) *http.Client {
	var rt http.RoundTripper
	if existingClient != nil && existingClient.Transport != nil {
		rt = existingClient.Transport
	}
	return &http.Client{
		Timeout:   time.Second * time.Duration(runtimeCtx.APIRequestTimeout),
		Transport: getRoundTripper(runtimeCtx, rt),
	}
}
