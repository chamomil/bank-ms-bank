package http

import (
	"net/http"
)

type (
	middleware      func(http.HandlerFunc) http.HandlerFunc
	middlewareGroup []middleware
)

func (mg middlewareGroup) Apply(h http.HandlerFunc) http.HandlerFunc {
	for i := len(mg) - 1; i >= 0; i-- {
		h = mg[i](h)
	}

	return h
}
