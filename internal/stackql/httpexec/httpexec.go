package httpexec

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql/internal/stackql/util"
	// log "github.com/sirupsen/logrus"
)

func getResponseMediaType(r *http.Response) (string, error) {
	rt := r.Header.Get("Content-Type")
	var mediaType string
	var err error
	if rt != "" {
		mediaType, _, err = mime.ParseMediaType(rt)
		if err != nil {
			return "", err
		}
		return mediaType, nil
	}
	return "", nil
}

func marshalResponse(r *http.Response) (interface{}, error) {
	body := r.Body
	if body != nil {
		defer body.Close()
	} else {
		return nil, nil
	}
	var target interface{}
	mediaType, err := getResponseMediaType(r)
	if err != nil {
		return nil, err
	}
	switch mediaType {
	case openapistackql.MediaTypeJson:
		err = json.NewDecoder(body).Decode(&target)
	case openapistackql.MediaTypeXML:
		err = xml.NewDecoder(body).Decode(&target)
	case openapistackql.MediaTypeOctetStream:
		target, err = io.ReadAll(body)
	case openapistackql.MediaTypeTextPlain, openapistackql.MediaTypeHTML:
		var b []byte
		b, err = io.ReadAll(body)
		if err == nil {
			target = string(b)
		}
	default:
		target, err = io.ReadAll(body)
	}
	return target, err
}

func ProcessHttpResponse(response *http.Response) (interface{}, error) {
	target, err := marshalResponse(response)
	if err == nil && response.StatusCode >= 400 {
		err = fmt.Errorf(fmt.Sprintf("HTTP response error: %s", string(util.InterfaceToBytes(target, true))))
	}
	if err == io.EOF {
		if response.StatusCode >= 200 && response.StatusCode < 300 {
			return map[string]interface{}{"result": "The Operation Completed Successfully"}, nil
		}
	}
	switch rv := target.(type) {
	case string, int:
		return map[string]interface{}{openapistackql.AnonymousColumnName: []interface{}{rv}}, nil
	}
	return target, err
}

func DeprecatedProcessHttpResponse(response *http.Response) (map[string]interface{}, error) {
	target, err := ProcessHttpResponse(response)
	if err != nil {
		return nil, err
	}
	switch rv := target.(type) {
	case map[string]interface{}:
		return rv, nil
	case nil:
		return nil, nil
	case string:
		return map[string]interface{}{openapistackql.AnonymousColumnName: rv}, nil
	case []byte:
		return map[string]interface{}{openapistackql.AnonymousColumnName: string(rv)}, nil
	default:
		return nil, fmt.Errorf("DeprecatedProcessHttpResponse() cannot acccept response of type %T", rv)
	}
}
