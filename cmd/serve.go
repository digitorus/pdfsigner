package cmd

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves the web API",
	Long:  `Long multiline description here`,
	Run: func(cmd *cobra.Command, args []string) {
		r := mux.NewRouter()
		r.HandleFunc("/put-job", handlePut).Methods("POST")
		r.HandleFunc("/get", handleGet).Methods("GET")
		r.HandleFunc("/check", handleCheck).Methods("GET")
	},
}

func handlePut(w http.ResponseWriter, r *http.Request) {

}
func handleCheck(w http.ResponseWriter, r *http.Request) {

}
func handleGet(w http.ResponseWriter, r *http.Request) {

}

func init() {
	RootCmd.AddCommand(serveCmd)
}
