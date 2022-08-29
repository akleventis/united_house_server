package lib

import (
	"encoding/json"
	"net/http"
)

func ApiResponse(w http.ResponseWriter, code int, obj interface{}) {
	r, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(r)
}
