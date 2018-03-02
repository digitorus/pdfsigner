package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func (wa *WebAPI) handleGetQueueSize(w http.ResponseWriter, r *http.Request) {
	// get jobs for session
	vars := mux.Vars(r)
	signerName := vars["signerName"]

	queue, err := wa.qSign.GetQueueSizeBySignerName(signerName)
	if err != nil {
		httpError(w, err, 500)
	}

	// respond with json
	j, err := json.Marshal(queue)
	if err != nil {
		httpError(w, err, 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
