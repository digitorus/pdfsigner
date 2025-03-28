package webapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// httpErr represents the error response to the user.
type httpErr struct {
	// Error represents error message
	Error string `json:"error"`
	// Code represents error code
	Code int `json:"code"`
}

// httpError writes to the response writer error and the code in json format.
func httpError(w http.ResponseWriter, err error, code int) error {
	e := httpErr{Error: errors.Cause(err).Error(), Code: code}

	// respond with json
	_ = respondJSON(w, e, code)

	return err
}

// respondJSON responds with json.
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
	_, err = w.Write(j)

	return err
}
