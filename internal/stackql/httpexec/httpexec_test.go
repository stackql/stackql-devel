package httpexec_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	. "github.com/stackql/stackql/internal/stackql/httpexec"
	"gotest.tools/assert"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestPlaceholder(t *testing.T) {
	res := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(`{"a": { "b": [ "c" ] } }`)),
	}
	_, err := ProcessHttpResponse(res)
	assert.NilError(t, err)
}
