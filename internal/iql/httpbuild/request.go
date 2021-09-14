package httpbuild

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"infraql/internal/iql/dto"
	"infraql/internal/iql/handler"
	"infraql/internal/iql/httpexec"
	"infraql/internal/iql/metadata"
	"infraql/internal/iql/parserutil"
	"infraql/internal/iql/provider"
	"infraql/internal/iql/requests"
	"infraql/internal/iql/util"

	"vitess.io/vitess/go/vt/sqlparser"

	log "github.com/sirupsen/logrus"
)

type ExecContext struct {
	ExecPayload *dto.ExecPayload
	Resource    *metadata.Resource
}

func NewExecContext(payload *dto.ExecPayload, rsc *metadata.Resource) *ExecContext {
	return &ExecContext{
		ExecPayload: payload,
		Resource:    rsc,
	}
}

type HTTPArmouryParameters struct {
	Header     http.Header
	Parameters *dto.HttpParameters
	Context    httpexec.IHttpContext
	BodyBytes  []byte
}

type HTTPArmoury struct {
	RequestParams  []HTTPArmouryParameters
	RequestSchema  *metadata.Schema
	ResponseSchema *metadata.Schema
}

func NewHTTPArmouryParameters() HTTPArmouryParameters {
	return HTTPArmouryParameters{
		Header: make(http.Header),
	}
}

func NewHTTPArmoury() HTTPArmoury {
	return HTTPArmoury{}
}

func BuildHTTPRequestCtx(handlerCtx *handler.HandlerContext, node sqlparser.SQLNode, prov provider.IProvider, m *metadata.Method, schemaMap map[string]metadata.Schema, insertValOnlyRows map[int]map[int]interface{}, execContext *ExecContext) (*HTTPArmoury, error) {
	var err error
	if m.Protocol != "http" {
		return nil, nil
	}
	httpArmoury := NewHTTPArmoury()
	requestSchema, ok := schemaMap[m.RequestType.Type]
	if ok {
		httpArmoury.RequestSchema = &requestSchema
	} else {
		log.Infoln(fmt.Sprintf("cannot locate schema for response type = '%s'", m.RequestType.Type))
	}
	responseSchema, ok := schemaMap[m.ResponseType.Type]
	if ok {
		httpArmoury.ResponseSchema = &responseSchema
	} else {
		log.Infoln(fmt.Sprintf("cannot locate schema for response type = '%s'", m.ResponseType.Type))
	}
	if err != nil {
		return nil, err
	}
	paramMap, err := util.ExtractSQLNodeParams(node, insertValOnlyRows)
	if err != nil {
		return nil, err
	}
	paramList, err := requests.SplitHttpParameters(prov, paramMap, m, httpArmoury.RequestSchema, httpArmoury.ResponseSchema)
	if err != nil {
		return nil, err
	}
	for _, params := range paramList {
		pm := NewHTTPArmouryParameters()
		if err != nil {
			return nil, err
		}
		if execContext != nil && execContext.ExecPayload != nil {
			pm.BodyBytes = execContext.ExecPayload.Payload
			for j, v := range execContext.ExecPayload.Header {
				pm.Header[j] = v
			}

		}
		if params.RequestBody != nil && len(params.RequestBody) != 0 {
			b, err := json.Marshal(params.RequestBody)
			if err != nil {
				return nil, err
			}
			pm.BodyBytes = b
			pm.Header["Content-Type"] = []string{"application/json"}
		}
		pm.Parameters = params
		httpArmoury.RequestParams = append(httpArmoury.RequestParams, pm)
	}
	var baseRequestCtx httpexec.IHttpContext
	switch node := node.(type) {
	case *sqlparser.Delete:
		baseRequestCtx, err = getDeleteRequestCtx(handlerCtx, prov, node, m)
	case *sqlparser.Exec:
		baseRequestCtx, err = getExecRequestCtx(execContext.Resource, m)
	case *sqlparser.Insert:
		baseRequestCtx, err = getInsertRequestCtx(handlerCtx, prov, node, m)
	case *sqlparser.Select:
		baseRequestCtx, err = getSelectRequestCtx(handlerCtx, prov, node, m)
	default:
		return nil, fmt.Errorf("cannot create http primitive for sql node of type %T", node)
	}
	if err != nil {
		return nil, err
	}
	for i, p := range httpArmoury.RequestParams {
		log.Infoln(fmt.Sprintf("pre transform: httpArmoury.RequestParams[%d] = %s", i, string(p.BodyBytes)))
		p.Context, err = prov.Parameterise(baseRequestCtx, m, p.Parameters, httpArmoury.RequestSchema)
		if handlerCtx.RuntimeContext.HTTPLogEnabled {
			url, _ := p.Context.GetUrl()
			handlerCtx.OutErrFile.Write([]byte(fmt.Sprintln(fmt.Sprintf("http request url: %s", url))))
		}
		if p.Header == nil {
			p.Header = make(http.Header)
		}
		p.Header["Content-Type"] = []string{"application/json"}
		if p.BodyBytes != nil && p.Header != nil && len(p.Header) > 0 {
			requestBodyMsg := fmt.Sprintf("http request body: %s", string(p.BodyBytes))
			log.Infoln(requestBodyMsg)
			if handlerCtx.RuntimeContext.HTTPLogEnabled {
				handlerCtx.OutErrFile.Write([]byte(fmt.Sprintln(requestBodyMsg)))
			}
			p.Context.SetBody(bytes.NewReader(p.BodyBytes))
			p.Context.SetHeaders(p.Header)
		}
		log.Infoln(fmt.Sprintf("post transform: httpArmoury.RequestParams[%d] = %s", i, string(p.BodyBytes)))
		httpArmoury.RequestParams[i] = p
	}
	if err != nil {
		return nil, err
	}
	return &httpArmoury, nil
}

func getSelectRequestCtx(handlerCtx *handler.HandlerContext, prov provider.IProvider, node *sqlparser.Select, method *metadata.Method) (httpexec.IHttpContext, error) {
	var path string
	var httpVerb string
	var err error
	currentSvcRsc, _ := parserutil.TableFromSelectNode(node)
	currentService := currentSvcRsc.Qualifier.GetRawVal()
	currentResource := currentSvcRsc.Name.GetRawVal()
	rsc, err := prov.GetResource(currentService, currentResource, handlerCtx.RuntimeContext)
	path = path + rsc.BaseUrl
	path = path + method.Path
	httpVerb = method.Verb
	return httpexec.CreateTemplatedHttpContext(
			httpVerb,
			path,
			nil,
		),
		err
}

func getDeleteRequestCtx(handlerCtx *handler.HandlerContext, prov provider.IProvider, node *sqlparser.Delete, method *metadata.Method) (httpexec.IHttpContext, error) {
	var path string
	var httpVerb string
	var err error
	currentSvcRsc, err := parserutil.ExtractSingleTableFromTableExprs(node.TableExprs)
	if err != nil {
		return nil, err
	}
	currentService := currentSvcRsc.Qualifier.GetRawVal()
	currentResource := currentSvcRsc.Name.GetRawVal()
	rsc, err := prov.GetResource(currentService, currentResource, handlerCtx.RuntimeContext)
	path = path + rsc.BaseUrl
	path = path + method.Path
	httpVerb = method.Verb
	return httpexec.CreateTemplatedHttpContext(
			httpVerb,
			path,
			nil,
		),
		err
}

func getInsertRequestCtx(handlerCtx *handler.HandlerContext, prov provider.IProvider, node *sqlparser.Insert, method *metadata.Method) (httpexec.IHttpContext, error) {
	var path string
	var httpVerb string
	var err error
	currentSvcRsc := node.Table
	currentService := currentSvcRsc.Qualifier.GetRawVal()
	currentResource := currentSvcRsc.Name.GetRawVal()
	rsc, err := prov.GetResource(currentService, currentResource, handlerCtx.RuntimeContext)
	path = path + rsc.BaseUrl
	path = path + method.Path
	httpVerb = method.Verb
	return httpexec.CreateTemplatedHttpContext(
			httpVerb,
			path,
			nil,
		),
		err
}

func getExecRequestCtx(rsc *metadata.Resource, method *metadata.Method) (httpexec.IHttpContext, error) {
	path := rsc.BaseUrl + method.Path
	httpVerb := method.Verb
	return httpexec.CreateTemplatedHttpContext(
			httpVerb,
			path,
			nil,
		),
		nil
}
