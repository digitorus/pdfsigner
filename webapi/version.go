package webapi

import (
	"net/http"
)

func (wa *WebAPI) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	// respond with json
	respondJSON(w, wa.version, http.StatusOK)
}
