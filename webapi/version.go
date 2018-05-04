package webapi

import (
	"net/http"
)

func (wa *WebAPI) handleGetVersion(w http.ResponseWriter, r *http.Request) error {
	// respond with json
	return respondJSON(w, wa.version, http.StatusOK)
}
