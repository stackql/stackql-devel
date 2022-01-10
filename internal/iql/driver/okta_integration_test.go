package driver_test

import (
	"bufio"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"testing"

	. "infraql/internal/iql/driver"
	"infraql/internal/iql/querysubmit"
	"infraql/internal/iql/responsehandler"
	"infraql/internal/iql/util"

	"infraql/internal/iql/config"
	"infraql/internal/iql/entryutil"
	"infraql/internal/iql/handler"
	"infraql/internal/iql/provider"

	"infraql/internal/test/infraqltestutil"
	"infraql/internal/test/testhttpapi"
	"infraql/internal/test/testobjects"

	lrucache "vitess.io/vitess/go/cache"
)

func TestSelectOktaApplicationAppsDriver(t *testing.T) {
	// SimpleOktaApplicationsAppsListResponseFile

	responseFile1, err := util.GetFilePathFromRepositoryRoot(testobjects.SimpleOktaApplicationsAppsListResponseFile)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	responseBytes1, err := ioutil.ReadFile(responseFile1)
	if err != nil {
		t.Fatalf("%v", err)
	}

	runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text", "TestSelectOktaApplicationAppsDriver")
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	path := "/api/v1/apps"
	url := &url.URL{
		Path: path,
	}
	ex := testhttpapi.NewHTTPRequestExpectations(nil, nil, "GET", url, "some-silly-subdomain.okta.com", string(responseBytes1), nil)
	expectations := map[string]testhttpapi.HTTPRequestExpectations{
		"some-silly-subdomain.okta.com" + path: *ex,
	}
	exp := testhttpapi.NewExpectationStore(1)
	for k, v := range expectations {
		exp.Put(k, v)
	}
	testhttpapi.StartServer(t, exp)
	provider.DummyAuth = true

	sqlEng, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	handlerCtx, err := handler.GetHandlerCtx(testobjects.SimpleSelectOktaApplicationApps, *runtimeCtx, lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), sqlEng)
	handlerCtx.Outfile = os.Stdout
	handlerCtx.OutErrFile = os.Stderr

	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	tc, err := entryutil.GetTxnCounterManager(handlerCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	handlerCtx.TxnCounterMgr = tc

	ProcessQuery(&handlerCtx)

	t.Logf("simple select driver integration test passed")
}

func TestSimpleSelectOktaApplicationAppsDriverOutput(t *testing.T) {
	runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text", "TestSimpleSelectOktaApplicationAppsDriverOutput")
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

		handlerCtx, err := entryutil.BuildHandlerContext(*runtimeCtx, strings.NewReader(""), lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), sqlEngine)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		handlerCtx.Outfile = outFile
		handlerCtx.OutErrFile = os.Stderr

		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx.Query = testobjects.SimpleSelectOktaApplicationApps
		response := querysubmit.SubmitQuery(&handlerCtx)
		handlerCtx.Outfile = outFile
		responsehandler.HandleResponse(&handlerCtx, response)
	}

	infraqltestutil.SetupSelectOktaApplicationApps(t)
	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedSelectOktaApplicationAppsJson})

}
