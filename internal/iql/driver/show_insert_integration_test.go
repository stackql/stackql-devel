package driver_test

import (
	"os"
	"testing"

	"bufio"

	. "infraql/internal/iql/driver"
	"infraql/internal/iql/util"

	"infraql/internal/iql/config"
	"infraql/internal/iql/entryutil"

	"infraql/internal/test/infraqltestutil"
	"infraql/internal/test/testobjects"

	lrucache "vitess.io/vitess/go/cache"
)

func TestSimpleShowInsertComputeAddressesRequired(t *testing.T) {

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

		runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text", "TestSimpleShowInsertComputeAddressesRequired")
		if err != nil {
			t.Fatalf("TestSimpleTemplateComputeAddressesRequired failed: %v", err)
		}
		sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		showInsertFile, err := util.GetFilePathFromRepositoryRoot(testobjects.ShowInsertAddressesRequiredInputFile)
		if err != nil {
			t.Fatalf("TestSimpleTemplateComputeAddressesRequired failed: %v", err)
		}
		runtimeCtx.InfilePath = showInsertFile
		runtimeCtx.CSVHeadersDisable = true

		rdr, err := os.Open(runtimeCtx.InfilePath)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx, err := entryutil.BuildHandlerContext(*runtimeCtx, rdr, lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), sqlEngine)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx.Outfile = outFile
		handlerCtx.OutErrFile = os.Stderr

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedShowInsertAddressesRequiredFile})

}

func TestSimpleShowInsertBiqueryDatasets(t *testing.T) {

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

		runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text", "TestSimpleShowInsertBiqueryDatasets")
		if err != nil {
			t.Fatalf("TestSimpleShowInsertBiqueryDatasets failed: %v", err)
		}
		sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		showInsertFile, err := util.GetFilePathFromRepositoryRoot(testobjects.ShowInsertBQDatasetsFile)
		if err != nil {
			t.Fatalf("TestSimpleShowInsertBiqueryDatasets failed: %v", err)
		}
		runtimeCtx.InfilePath = showInsertFile
		runtimeCtx.CSVHeadersDisable = true

		rdr, err := os.Open(runtimeCtx.InfilePath)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx, err := entryutil.BuildHandlerContext(*runtimeCtx, rdr, lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), sqlEngine)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx.Outfile = outFile
		handlerCtx.OutErrFile = os.Stderr

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedShowInsertBQDatasetsFile})

}

func TestSimpleShowInsertBiqueryDatasetsRequired(t *testing.T) {

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

		runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text", "TestSimpleShowInsertBiqueryDatasetsRequired")
		if err != nil {
			t.Fatalf("TestSimpleShowInsertBiqueryDatasetsRequired failed: %v", err)
		}
		sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		showInsertFile, err := util.GetFilePathFromRepositoryRoot(testobjects.ShowInsertBQDatasetsRequiredFile)
		if err != nil {
			t.Fatalf("TestSimpleShowInsertBiqueryDatasetsRequired failed: %v", err)
		}
		runtimeCtx.InfilePath = showInsertFile
		runtimeCtx.CSVHeadersDisable = true

		rdr, err := os.Open(runtimeCtx.InfilePath)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx, err := entryutil.BuildHandlerContext(*runtimeCtx, rdr, lrucache.NewLRUCache(int64(runtimeCtx.QueryCacheSize)), sqlEngine)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		handlerCtx.Outfile = outFile
		handlerCtx.OutErrFile = os.Stderr

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedShowInsertBQDatasetsRequiredFile})

}
