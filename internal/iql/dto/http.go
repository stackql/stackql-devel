package dto

type HttpParameters struct {
	PathParams   map[string]interface{}
	QueryParams  map[string]interface{}
	RequestBody  map[string]interface{}
	ResponseBody map[string]interface{}
	Unassigned   map[string]interface{}
}

func NewHttpParameters() *HttpParameters {
	return &HttpParameters{
		PathParams:   make(map[string]interface{}),
		QueryParams:  make(map[string]interface{}),
		RequestBody:  make(map[string]interface{}),
		ResponseBody: make(map[string]interface{}),
		Unassigned:   make(map[string]interface{}),
	}
}
