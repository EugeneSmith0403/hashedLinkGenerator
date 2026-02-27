package response

import (
	"encoding/json"
	"net/http"
)

type ResponseOptions struct {
	HeadersMap map[string]string
}

type Response struct {
	Options *ResponseOptions
}

type JsonOptions struct {
	Data   any
	Code   int
	Writer http.ResponseWriter
	Reader *http.Request
}

func NewResponse(options *ResponseOptions) *Response {
	return &Response{
		Options: options,
	}
}

func (r *Response) Json(options *JsonOptions) {
	r.setHeaders(r.Options.HeadersMap, options.Writer)
	options.Writer.WriteHeader(options.Code)
	json.NewEncoder(options.Writer).Encode(options.Data)
}

func (r *Response) setHeaders(headerMap map[string]string, writer http.ResponseWriter) {
	for key, value := range headerMap {
		writer.Header().Set(key, value)
	}
}
