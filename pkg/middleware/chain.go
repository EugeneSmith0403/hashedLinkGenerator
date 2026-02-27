package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func Chain(middlewares ...Middleware) Middleware {

	return func(next http.Handler) http.Handler {

		for _, v := range middlewares {

			next = v(next)

		}

		return next

	}

}
