package request

import (
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/response"
	"encoding/json"
	"io"
)

type DecoreOptions[T any] struct {
	Body        io.ReadCloser
	Res         response.Response
	JsonOptions *response.JsonOptions
	Payload     *T
}

func Decoder[T any](options DecoreOptions[T]) {
	err := json.NewDecoder(options.Body).Decode(&options.Payload)

	if err != nil {
		options.JsonOptions.Data = &errorType.ErrorType{
			Error: err.Error(),
		}
		options.Res.Json(options.JsonOptions)
	}
}
