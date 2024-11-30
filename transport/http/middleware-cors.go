package http

import "net/http"

func (t *Transport) corsMiddleware(cors http.HandlerFunc) middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cors(w, r)
			handlerFunc(w, r)
		}
	}
}
