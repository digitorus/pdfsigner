package webapi

import (
	"encoding/json"
	"net/http"
)

func (wa *WebAPI) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	// respond with json
	j, err := json.Marshal(wa.version)
	if err != nil {
		httpError(w, err, 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
