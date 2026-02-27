package request

import (
	"adv/go-http/pkg/response"
	"net/http"
)

func HandleBody[T any](req *http.Request, w http.ResponseWriter, res response.Response) (T, error) {

	var payload T
	var noResult T

	Decoder(DecoreOptions[T]{
		Res:     res,
		Body:    req.Body,
		Payload: &payload,
		JsonOptions: &response.JsonOptions{
			Code:   500,
			Writer: w,
			Reader: req,
		},
	})

	_, err := isValid(res, &response.JsonOptions{
		Code:   423,
		Writer: w,
		Reader: req,
	}, &payload)

	if err != nil {
		return noResult, err
	}

	return payload, nil

}
