package webapi

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"bitbucket.org/digitorus/pdfsigner/queues/queue"
	"bitbucket.org/digitorus/pdfsigner/version"
	"github.com/gorilla/mux"
)

// handler represents mux handle function
type handler func(w http.ResponseWriter, r *http.Request) error

// middleware represents middleware that could be added to handle function
type middleware func(handler) handler

// WebAPI represents all the data related to webapi
type WebAPI struct {
	// r represents router
	r *mux.Router
	// addr represents address
	addr string
	// queue represents sign queue
	queue *queue.Queue
	// allowedUnits represents signers that allowed to be used by the web api
	allowedUnits []string
	// version represents git version of the application
	version version.Version
	// middlewares represents middlewares used for all handlers
	middlewares []middleware
	// defaultValidateSignature defines defaults for signature validation after signing
	defaultValidateSignature bool
}

// NewWebAPI initializes web api with routes
func NewWebAPI(addr string, qs *queue.Queue, allowedUnits []string, version version.Version, defaultValidateSignature bool) *WebAPI {

	// initialize web api
	wa := WebAPI{
		addr:                     addr,
		queue:                    qs,
		allowedUnits:             allowedUnits,
		version:                  version,
		r:                        mux.NewRouter(),
		middlewares:              []middleware{},
		defaultValidateSignature: defaultValidateSignature,
	}

	wa.allowedUnits = append(allowedUnits, "verify")

	// add middlewares
	wa.addMiddleware(loggerMiddleware)
	wa.addMiddleware(errorHandler)

	// initialize sign routes
	wa.handle("POST", "/sign", wa.handleSignSchedule)
	wa.handle("GET", "/sign/{jobID}", wa.handleStatus)
	wa.handle("GET", "/sign/{jobID}/{taskID}/download", wa.handleSignGetFile)
	wa.handle("DELETE", "/sign/{jobID}", wa.handleDelete)
	wa.handle("GET", "/queue/{unitName}", wa.handleGetQueueSize)
	wa.handle("GET", "/version", wa.handleGetVersion)

	// initialize verify routes
	wa.handle("POST", "/verify", wa.handleVerifySchedule)
	wa.handle("GET", "/verify/{jobID}", wa.handleStatus)
	wa.handle("GET", "/verify/{jobID}/info/{taskID}", wa.handleVerifyGetInfo)

	return &wa
}

// handle adds middlewares and runs mux handler
func (wa *WebAPI) handle(verb string, path string, handler handler) {
	// create handler function
	h := func(w http.ResponseWriter, r *http.Request) {

		// add middlewares
		for i := len(wa.middlewares) - 1; i >= 0; i-- {
			if wa.middlewares[i] != nil {
				handler = wa.middlewares[i](handler)
			}
		}

		// add handler
		handler(w, r)
	}

	// run result handler function
	wa.r.HandleFunc(path, h).Methods(verb)
}

// addMiddleware adds middleware to web api to be  run before handler
func (wa *WebAPI) addMiddleware(m middleware) {
	wa.middlewares = append(wa.middlewares, m)
}

// Serve starts the web server
func (wa *WebAPI) Serve() {
	// create server
	s := &http.Server{
		Addr:           wa.addr,
		Handler:        wa.r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serveLoggerCtx := log.WithFields(log.Fields{
		"addr":         wa.addr,
		"allowedUnits": wa.allowedUnits,
	})
	serveLoggerCtx.Info("Starting Web API...")

	if err := s.ListenAndServe(); err != nil {
		serveLoggerCtx.Fatal("Coudn't start Web API:", err)
	}
}
