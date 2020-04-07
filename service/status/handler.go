package status

import (
	"encoding/json"
	"net/http"

	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/CanonicalLtd/serial-vault/service/response"
	"github.com/gorilla/mux"
)

// Handler returns an http.Handler for all /_status/ routes
func AddStatusEndpoints(prefix string, r *mux.Router) {
	s := r.PathPrefix(prefix).Subrouter()

	s.HandleFunc("/ping", PingHandler).
		Methods("GET")

	s.HandleFunc("/check", DatabasePingHandler).
		Methods("GET")
}

// PingHandler returns 200 OK response with version of the service in the body
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(datastore.Environ.Config.Version))
}

// DatabasePingHandler will return a json data with
// 200: { "database": "OK" }
// or
// 500: { "database": "dial tcp 127.0.0.1:5432: connect: connection refused" }
func DatabasePingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", response.JSONHeader)
	status := "OK"

	err := datastore.Environ.DB.HealthCheck()
	if err != nil {
		status = err.Error()
		w.WriteHeader(500)
	}

	json.NewEncoder(w).Encode(map[string]string{"database": status})
}
