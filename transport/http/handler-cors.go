package http

import "net/http"

func (t *Transport) corsHandler(allowOrigin, allowHeaders, allowMethods, exposeHeaders string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		}
		if allowMethods != "" {
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		}
		if allowMethods != "" {
			w.Header().Set("Access-Control-Allow-Methods", allowMethods)
		}
		if exposeHeaders != "" {
			w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
		}
	}
}
