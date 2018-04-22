package webapi

import (
	"log"
	"net/http"
	"time"

	"bitbucket.org/digitorus/pdfsigner/queues/queue"
	"bitbucket.org/digitorus/pdfsigner/version"
	"github.com/gorilla/mux"
)

// WebAPI represents all the data related to webapi
type WebAPI struct {
	// r represents router
	r *mux.Router
	// addr represents address
	addr string
	// queue represents sign queue
	queue *queue.Queue
	// allowedSigners represents signers that allowed to be used by the web api
	allowedSigners []string
	// version represents git version of the application
	version version.Version
}

// NewWebAPI initializes web api with routes
func NewWebAPI(addr string, qs *queue.Queue, allowedSigners []string, version version.Version) *WebAPI {
	wa := WebAPI{
		addr:           addr,
		queue:          qs,
		allowedSigners: allowedSigners,
		version:        version,
		r:              mux.NewRouter(),
	}

	// initialize sign routes
	wa.r.HandleFunc("/sign", wa.handleSignSchedule).Methods("POST")
	wa.r.HandleFunc("/sign/{jobID}", wa.handleSignStatus).Methods("GET")
	wa.r.HandleFunc("/sign/{jobID}/{taskID}/download", wa.handleSignGetFile).Methods("GET")
	wa.r.HandleFunc("/sign/{jobID}", wa.handleSignDelete).Methods("DELETE")
	wa.r.HandleFunc("/queue/{signerName}", wa.handleGetQueueSize).Methods("GET")
	wa.r.HandleFunc("/version", wa.handleGetVersion).Methods("GET")

	// initialize verify routes
	wa.r.HandleFunc("/verify/schedule", wa.handleVerifySchedule).Methods("POST")
	wa.r.HandleFunc("/verify/check", wa.handleVerifyCheck).Methods("POST")

	return &wa
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

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Coudn't start Web API:", err)
	}
}
