package webapi

import (
	"log"
	"net/http"
	"time"

	"bitbucket.org/digitorus/pdfsigner/queued_sign"
	"bitbucket.org/digitorus/pdfsigner/queued_verify"
	"bitbucket.org/digitorus/pdfsigner/version"
	"github.com/gorilla/mux"
)

type WebAPI struct {
	r              *mux.Router
	addr           string
	qSign          *queued_sign.QSign
	qVerify        *queued_verify.QVerify
	allowedSigners []string
	version        version.Version
}

func NewWebAPI(addr string, qs *queued_sign.QSign, qv *queued_verify.QVerify, allowedSigners []string, version version.Version) *WebAPI {
	wa := WebAPI{
		addr:           addr,
		qSign:          qs,
		qVerify:        qv,
		allowedSigners: allowedSigners,
		version:        version,
		r:              mux.NewRouter(),
	}

	// sign
	wa.r.HandleFunc("/sign", wa.handleSignSchedule).Methods("POST")
	wa.r.HandleFunc("/sign/{jobID}", wa.handleSignCheck).Methods("GET")
	wa.r.HandleFunc("/sign/{jobID}/{taskID}/download", wa.handleSignGetFile).Methods("GET")
	wa.r.HandleFunc("/sign/{jobID}", wa.handleSignDelete).Methods("DELETE")
	wa.r.HandleFunc("/queue/{signerName}", wa.handleGetQueueSize).Methods("GET")
	wa.r.HandleFunc("/version", wa.handleGetVersion).Methods("GET")

	//verify
	wa.r.HandleFunc("/verify/schedule", wa.handleVerifySchedule).Methods("POST")
	wa.r.HandleFunc("/verify/check", wa.handleVerifyCheck).Methods("POST")

	return &wa
}

func (wa *WebAPI) Serve() {
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
