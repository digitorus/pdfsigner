package webapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

// handleGetQueueSize responses with queue size by signer name
func (wa *WebAPI) handleGetQueueSize(w http.ResponseWriter, r *http.Request) {
	// get tasks for job
	vars := mux.Vars(r)
	signerName := vars["signerName"]

	// get queue sizes by signer name
	queue, err := wa.queue.GetQueueSizeByUnitName(signerName)
	if err != nil {
		httpError(w, err, 500)
	}

	respondJSON(w, queue, http.StatusOK)
}
