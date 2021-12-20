package handler

import (
	"fmt"
	"infraql/internal/iql/drm"
	"infraql/internal/iql/dto"
	"infraql/internal/iql/provider"
	"infraql/internal/iql/sqlengine"
	"infraql/internal/pkg/txncounter"
	"io"

	lrucache "vitess.io/vitess/go/cache"
)

var (
	drmConfig drm.DRMConfig = drm.GetGoogleV1SQLiteConfig()
)

type HandlerContext struct {
	RawQuery          string
	Query             string
	RuntimeContext    dto.RuntimeCtx
	providers         map[string]provider.IProvider
	CurrentProvider   string
	authContexts      map[string]*dto.AuthCtx
	ErrorPresentation string
	Outfile           io.Writer
	OutErrFile        io.Writer
	LRUCache          *lrucache.LRUCache
	SQLEngine         sqlengine.SQLEngine
	DrmConfig         drm.DRMConfig
	TxnCounterMgr     *txncounter.TxnCounterManager
}

func (hc *HandlerContext) GetProvider(providerName string) (provider.IProvider, error) {
	var err error
	if providerName == "" {
		providerName = hc.RuntimeContext.ProviderStr
	}
	prov, ok := hc.providers[providerName]
	if !ok {
		prov, err = provider.GetProvider(hc.RuntimeContext, providerName, "v1", hc.SQLEngine)
		if err == nil {
			hc.providers[providerName] = prov
			return prov, err
		}
		err = fmt.Errorf("cannot find provider = '%s': %s", providerName, err.Error())
	}
	return prov, err
}

func (hc *HandlerContext) GetAuthContext(providerName string) (*dto.AuthCtx, error) {
	var err error
	if providerName == "" {
		providerName = hc.RuntimeContext.ProviderStr
	}
	authCtx, ok := hc.authContexts[providerName]
	if !ok {
		err = fmt.Errorf("cannot find AUTH context for provider = '%s'", providerName)
	}
	return authCtx, err
}

func GetHandlerCtx(cmdString string, runtimeCtx dto.RuntimeCtx, lruCache *lrucache.LRUCache, sqlEng sqlengine.SQLEngine) (HandlerContext, error) {
	prov, err := provider.GetProviderFromRuntimeCtx(runtimeCtx, sqlEng)
	if err != nil {
		return HandlerContext{}, err
	}
	return HandlerContext{
		RawQuery:       cmdString,
		RuntimeContext: runtimeCtx,
		providers: map[string]provider.IProvider{
			runtimeCtx.ProviderStr: prov,
		},
		authContexts: map[string]*dto.AuthCtx{
			runtimeCtx.ProviderStr: dto.GetAuthCtx(nil, runtimeCtx.KeyFilePath, runtimeCtx.KeyFileType),
		},
		ErrorPresentation: runtimeCtx.ErrorPresentation,
		LRUCache:          lruCache,
		SQLEngine:         sqlEng,
		DrmConfig:         drmConfig,
		TxnCounterMgr:     nil,
	}, nil
}
