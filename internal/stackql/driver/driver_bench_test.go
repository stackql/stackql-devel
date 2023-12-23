package driver_test

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	. "github.com/stackql/stackql/internal/stackql/driver"

	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/provider"

	"github.com/stackql/stackql/internal/test/stackqltestutil"
	"github.com/stackql/stackql/internal/test/testhttpapi"
	"github.com/stackql/stackql/internal/test/testobjects"

	lrucache "github.com/stackql/stackql-parser/go/cache"
)

//nolint:lll // legacy test
func BenchmarkSelectGoogleComputeInstanceDriver(b *testing.B) {
	runtimeCtx, err := stackqltestutil.GetRuntimeCtx(testobjects.GetGoogleProviderString(), "text", "BenchmarkSelectGoogleComputeInstanceDriver")
	if err != nil {
		b.Fatalf("Test failed: %v", err)
	}
	path := "/compute/v1/projects/testing-project/zones/australia-southeast1-b/instances"
	url := &url.URL{
		Path: path,
	}
	ex := testhttpapi.NewHTTPRequestExpectations(nil, nil, "GET", url, "compute.googleapis.com", testobjects.SimpleSelectGoogleComputeInstanceResponse, nil)
	expectations := map[string]testhttpapi.HTTPRequestExpectations{
		"compute.googleapis.com" + path: ex,
	}
	exp := testhttpapi.NewExpectationStore(1)
	for k, v := range expectations {
		exp.Put(k, v)
	}
	testhttpapi.StartServer(b, exp)
	provider.DummyAuth = true

	inputBundle, err := stackqltestutil.BuildInputBundle(*runtimeCtx)
	if err != nil {
		b.Fatalf("Test failed: %v", err)
	}

	stringQuery := `select name, zone from google.compute.instances where zone = 'australia-southeast1-b' AND /* */ project = 'testing-project';`

	handlerCtx, err := handler.GetHandlerCtx(stringQuery, *runtimeCtx, lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), inputBundle)
	handlerCtx.SetOutfile(os.Stdout)
	handlerCtx.SetOutErrFile(os.Stderr)

	dr, _ := NewStackQLDriver(handlerCtx)

	if err != nil {
		b.Fatalf("Test failed: %v", err)
	}

	dr.ProcessQuery(handlerCtx.GetRawQuery())

	b.Logf("benchmark select driver integration test passed")
}

//nolint:lll // legacy test
func BenchmarkParallelProjectSelectGoogleComputeInstanceDriver(b *testing.B) {
	runtimeCtx, err := stackqltestutil.GetRuntimeCtx(testobjects.GetGoogleProviderString(), "text", "BenchmarkParallelProjectSelectGoogleComputeInstanceDriver")
	if err != nil {
		b.Fatalf("Test failed: %v", err)
	}
	path := `/compute/v1/projects/%s/zones/australia-southeast1-b/instances`
	pathOne := fmt.Sprintf(path, "testing-project")
	urlOne := &url.URL{
		Path: pathOne,
	}
	exOne := testhttpapi.NewHTTPRequestExpectations(nil, nil, "GET", urlOne, "compute.googleapis.com", testobjects.SimpleSelectGoogleComputeInstanceResponse, nil)
	pathTwo := fmt.Sprintf(path, "testing-project-two")
	urlTwo := &url.URL{
		Path: pathTwo,
	}
	exTwo := testhttpapi.NewHTTPRequestExpectations(nil, nil, "GET", urlTwo, "compute.googleapis.com", testobjects.SimpleSelectGoogleComputeInstanceResponse, nil)
	expectations := map[string]testhttpapi.HTTPRequestExpectations{
		"compute.googleapis.com" + pathOne: exOne,
		"compute.googleapis.com" + pathTwo: exTwo,
	}
	exp := testhttpapi.NewExpectationStore(2)
	for k, v := range expectations {
		exp.Put(k, v)
	}
	testhttpapi.StartServer(b, exp)
	provider.DummyAuth = true

	inputBundle, err := stackqltestutil.BuildInputBundle(*runtimeCtx)
	if err != nil {
		b.Fatalf("Test failed: %v", err)
	}

	stringQuery := `select count(1) as inst_count from google.compute.instances where zone = 'australia-southeast1-b' AND /* */ project in ('testing-project', 'testing-project-two');`

	handlerCtx, err := handler.GetHandlerCtx(stringQuery, *runtimeCtx, lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), inputBundle)
	handlerCtx.SetOutfile(os.Stdout)
	handlerCtx.SetOutErrFile(os.Stderr)

	dr, _ := NewStackQLDriver(handlerCtx)

	if err != nil {
		b.Fatalf("Test failed: %v", err)
	}

	dr.ProcessQuery(handlerCtx.GetRawQuery())

	b.Logf("benchmark parallel select driver integration test passed")
}
