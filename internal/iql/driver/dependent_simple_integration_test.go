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

func TestSimpleInsertDependentGoogleComputeDiskAsync(t *testing.T) {
	runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text")
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	inputFile, err := util.GetFilePathFromRepositoryRoot(testobjects.SimpleInsertDependentComputeDisksFile)
	if err != nil {
		t.Fatalf("TestSimpleInsertDependentGoogleComputeNetworkAsync failed: %v", err)
	}
	runtimeCtx.InfilePath = inputFile
	sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

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

		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.SetupDependentInsertGoogleComputeDisks(t)
	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedComputeDisksDependentInsertAsyncFile})

}

func TestSimpleInsertDependentGoogleComputeDiskAsyncReversed(t *testing.T) {
	runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text")
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	inputFile, err := util.GetFilePathFromRepositoryRoot(testobjects.SimpleInsertDependentComputeDisksReversedFile)
	if err != nil {
		t.Fatalf("TestSimpleInsertDependentGoogleComputeNetworkAsync failed: %v", err)
	}
	runtimeCtx.InfilePath = inputFile
	sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

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

		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.SetupDependentInsertGoogleComputeDisks(t)
	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedComputeDisksDependentInsertAsyncFile})

}

func TestSimpleInsertDependentGoogleBQDatasetAsync(t *testing.T) {
	runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "text")
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	inputFile, err := util.GetFilePathFromRepositoryRoot(testobjects.SimpleInsertDependentBQDatasetFile)
	if err != nil {
		t.Fatalf("TestSimpleInsertDependentGoogleComputeNetworkAsync failed: %v", err)
	}
	runtimeCtx.InfilePath = inputFile
	sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

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

		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.SetupDependentInsertGoogleBQDatasets(t)
	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedBQDatasetsDependentInsertFile})

}

func TestSimpleSelectExecDependentGoogleOrganizationsGetIamPolicy(t *testing.T) {
	runtimeCtx, err := infraqltestutil.GetRuntimeCtx(config.GetGoogleProviderString(), "csv")
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}
	inputFile, err := util.GetFilePathFromRepositoryRoot(testobjects.SimpleSelectExecDependentOrgIamPolicyFile)
	if err != nil {
		t.Fatalf("TestSimpleSelectExecDependentGoogleOrganizationsGetIamPolicy failed: %v", err)
	}
	runtimeCtx.InfilePath = inputFile
	sqlEngine, err := infraqltestutil.BuildSQLEngine(*runtimeCtx)
	if err != nil {
		t.Fatalf("Test failed: %v", err)
	}

	testSubject := func(t *testing.T, outFile *bufio.Writer) {

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

		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}

		tc, err := entryutil.GetTxnCounterManager(handlerCtx)
		if err != nil {
			t.Fatalf("Test failed: %v", err)
		}
		handlerCtx.TxnCounterMgr = tc

		ProcessQuery(&handlerCtx)
	}

	infraqltestutil.SetupExecGoogleOrganizationsGetIamPolicy(t)
	infraqltestutil.RunCaptureTestAgainstFiles(t, testSubject, []string{testobjects.ExpectedSelectExecOrgGetIamPolicyAgg})

}
