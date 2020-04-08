package status

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CanonicalLtd/serial-vault/config"
	"github.com/CanonicalLtd/serial-vault/datastore"
	"github.com/gorilla/mux"
)

const version = "1.2.3"

func TestAddStatusEndpointsPing(t *testing.T) {
	// Mock the database
	config := config.Settings{Version: version}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}

	// run the test
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

func TestAddStatusEndpointsDBPingError(t *testing.T) {
	// Mock the database
	config := config.Settings{Version: version}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.DB = &datastore.ErrorMockDB{}

	// run the test
	router := mux.NewRouter()
	AddStatusEndpoints("/_status", router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_status/check", nil)

	router.ServeHTTP(w, r)

	if w.Code != 500 {
		t.Errorf("expected code 500, got %d", w.Code)
	}

	expected := `{"database":"Health check failed"}`
	got := strings.TrimSpace(w.Body.String())
	if expected != got {
		t.Errorf("expected body %s, got %s", expected, got)
	}
}

func TestAddStatusEndpointsDBPingOK(t *testing.T) {
	// Mock the database
	config := config.Settings{Version: version}
	datastore.Environ = &datastore.Env{DB: &datastore.MockDB{}, Config: config}
	datastore.Environ.DB = &datastore.MockDB{}

	// run the test
	router := mux.NewRouter()
	AddStatusEndpoints("/_status", router)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_status/check", nil)

	router.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected code 200, got %d", w.Code)
	}

	expected := `{"database":"OK"}`
	got := strings.TrimSpace(w.Body.String())
	if expected != got {
		t.Errorf("expected body %s, got %s", expected, got)
	}
}
