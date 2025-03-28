package webapi

import (
	"errors"
	"net/http"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

// loggerMiddleware logs the requests.
func loggerMiddleware(next handler) handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.WithFields(log.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote-addr": r.RemoteAddr,
		}).Info()

		// run next handler
		return next(w, r)
	}
}

func errorHandler(next handler) handler {
	h := func(w http.ResponseWriter, r *http.Request) error {
		// log error if panic
		defer func() {
			if r := recover(); r != nil {
				_ = httpError(w, errors.New("unhandled"), http.StatusInternalServerError)

				log.Print(debug.Stack())
			}
		}()

		// run next handler and log errors
		if err := next(w, r); err != nil {
			// log error
			log.WithFields(log.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote-addr": r.RemoteAddr,
			}).Error(err)

			return nil
		}

		return nil
	}

	return h
}
