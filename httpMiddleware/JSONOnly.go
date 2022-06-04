package httpMiddleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

func JSONOnly(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only write JSON
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Accept", "application/json; charset=utf-8")
		// only accept JSON requests
		receivedContentType := r.Header.Get("content-type")
		if r.Method != "GET" && strings.Contains(receivedContentType, "application/json") != true {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Reason": "This API only speaks in JSON!",
			})
			return
		}
		// serve inner
		h.ServeHTTP(w, r)
	})
}
