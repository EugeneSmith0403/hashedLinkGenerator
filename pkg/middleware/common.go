package middleware

import "net/http"

type WrapperWrite struct {
	http.ResponseWriter
	statusCode int
}

func (w *WrapperWrite) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
}
