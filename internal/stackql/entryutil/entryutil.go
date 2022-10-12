package entryutil

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/stackql/stackql/internal/stackql/dto"
	"github.com/stackql/stackql/internal/stackql/gc"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/iqlerror"
	"github.com/stackql/stackql/internal/stackql/sqlengine"

	"github.com/stackql/stackql/pkg/preprocessor"
	"github.com/stackql/stackql/pkg/txncounter"

	lrucache "vitess.io/vitess/go/cache"
)

func BuildSQLEngineAndGC(runtimeCtx dto.RuntimeCtx) (sqlengine.SQLEngine, gc.GarbageCollector, error) {
	se, err := buildSQLEngine(runtimeCtx)
	if err != nil {
		return nil, nil, err
	}
	gc, err := buildGC(se, runtimeCtx)
	if err != nil {
		return nil, nil, err
	}
	return se, gc, nil
}

func buildSQLEngine(runtimeCtx dto.RuntimeCtx) (sqlengine.SQLEngine, error) {
	sqlCfg := sqlengine.NewSQLEngineConfig(runtimeCtx)
	return sqlengine.NewSQLEngine(sqlCfg)
}

func buildGC(sqlEngine sqlengine.SQLEngine, runtimeCtx dto.RuntimeCtx) (gc.GarbageCollector, error) {
	return gc.NewGarbageCollector(sqlEngine, "sqlite")
}

func GetTxnCounterManager(handlerCtx handler.HandlerContext) (*txncounter.TxnCounterManager, error) {
	genId, err := handlerCtx.SQLEngine.GetCurrentGenerationId()
	if err != nil {
		genId, err = handlerCtx.SQLEngine.GetNextGenerationId()
		if err != nil {
			return nil, err
		}
	}
	sessionId, err := handlerCtx.SQLEngine.GetNextSessionId(genId)
	if err != nil {
		return nil, err
	}
	return txncounter.NewTxnCounterManager(genId, sessionId), nil
}

func PreprocessInline(runtimeCtx dto.RuntimeCtx, s string) (string, error) {
	rdr := strings.NewReader(s)
	bt, err := assemblePreprocessor(runtimeCtx, rdr)
	if err != nil || bt == nil {
		return s, err
	}
	return string(bt), nil
}

func assemblePreprocessor(runtimeCtx dto.RuntimeCtx, rdr io.Reader) ([]byte, error) {
	var err error
	var prepRd, externalTmplRdr io.Reader
	pp := preprocessor.NewPreprocessor(preprocessor.TripleLessThanToken, preprocessor.TripleGreaterThanToken)
	if pp == nil {
		return nil, fmt.Errorf("preprocessor error")
	}
	if runtimeCtx.TemplateCtxFilePath == "" {
		prepRd, err = pp.Prepare(rdr, runtimeCtx.InfilePath)
		if err != nil {
			return nil, err
		}
	} else {
		externalTmplRdr, err = os.Open(runtimeCtx.TemplateCtxFilePath)
		if err != nil {
			return nil, err
		}
		prepRd = rdr
		err = pp.PrepareExternal(strings.Trim(strings.ToLower(filepath.Ext(runtimeCtx.TemplateCtxFilePath)), "."), externalTmplRdr, runtimeCtx.TemplateCtxFilePath)
	}
	if err != nil {
		return nil, err
	}
	ppRd, err := pp.Render(prepRd)
	if err != nil {
		return nil, err
	}
	var bb []byte
	bb, err = ioutil.ReadAll(ppRd)
	return bb, err
}

func BuildHandlerContext(runtimeCtx dto.RuntimeCtx, rdr io.Reader, lruCache *lrucache.LRUCache, sqlEngine sqlengine.SQLEngine, garbageCollector gc.GarbageCollector) (handler.HandlerContext, error) {
	bb, err := assemblePreprocessor(runtimeCtx, rdr)
	iqlerror.PrintErrorAndExitOneIfError(err)
	return handler.GetHandlerCtx(strings.TrimSpace(string(bb)), runtimeCtx, lruCache, sqlEngine, garbageCollector)
}

func BuildHandlerContextNoPreProcess(runtimeCtx dto.RuntimeCtx, lruCache *lrucache.LRUCache, sqlEngine sqlengine.SQLEngine, garbageCollector gc.GarbageCollector) (handler.HandlerContext, error) {
	return handler.GetHandlerCtx("", runtimeCtx, lruCache, sqlEngine, garbageCollector)
}
