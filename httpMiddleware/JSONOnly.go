package httpMiddleware

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func JSONOnly(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// only write JSON
		w.Header().Set("Content-Type", "application/json")
		// only accept JSON requests
		receivedContentType := r.Header.Get("Content-Type")
		if receivedContentType != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Reason": "This API only accepts and conveys JSON data!",
			})
			return
		}
		// serve inner
		handle(w, r, p)
	}
}
