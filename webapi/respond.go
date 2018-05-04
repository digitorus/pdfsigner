package webapi

import (
	"encoding/json"
	"net/http"
)

// httpErr represents the error response to the user
type httpErr struct {
	// Message represents error message
	Message string `json:"message"`
	// Code represents error code
	Code int `json:"code"`
}

// httpError writes to the response writer error and the code in json format
func httpError(w http.ResponseWriter, err error, code int) error {
	e := httpErr{Message: err.Error(), Code: code}

	// respond with json
	respondJSON(w, e, code)

	return err
}

// respondJSON responds with json
func respondJSON(w http.ResponseWriter, data interface{}, code int) error {
	// marshal data
	j, err := json.Marshal(data)
	if err != nil {
		return httpError(w, err, http.StatusInternalServerError)
	}

	// set response code
	w.WriteHeader(code)

	// set content type
	w.Header().Set("Content-Type", "application/json")

	// respond with json
	w.Write(j)

	return nil
}
