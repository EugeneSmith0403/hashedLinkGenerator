package request

import (
	errorType "adv/go-http/pkg/errorType"
	"adv/go-http/pkg/response"

	"github.com/go-playground/validator/v10"
)

func isValid[T any](res response.Response, jsonOptionErr *response.JsonOptions, payload T) (bool, error) {
	validate := validator.New()
	valErr := validate.Struct(payload)

	if valErr != nil {
		jsonOptionErr.Data = &errorType.ErrorType{
			Error: valErr.Error(),
		}

		res.Json(jsonOptionErr)
		return false, valErr
	}

	return true, nil
}
