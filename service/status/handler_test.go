package status

import (
	"net/http/httptest"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

const version = "1.2.3"

func init() {
	// Mock the database
	config := config.Settings{Version: version}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
}

func TestAddStatusEndpointsPing(t *testing.T) {
	router := mux.NewRouter()
	AddStatusEndpoints("/_status", router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_status/ping", nil)

	router.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected code 200, got %d", w.Code)
	}
	if w.Body.String() != version {
		t.Errorf("expected body %s, got %s", version, w.Body.String())
	}
}
