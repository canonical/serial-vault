package service

import (
	"log"
	"net/http"
	"time"
)

// Logger Handle logging for the web service
func Logger(start time.Time, r *http.Request) {
	log.Printf(
		"%s\t%s\t%s",
		r.Method,
		r.RequestURI,
		time.Since(start),
	)
}

// Config contains the parsed config file settings.
var Config *ConfigSettings

// Middleware to pre-process web service requests
func Middleware(inner http.Handler, config *ConfigSettings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if Config == nil {
			Config = config
		}

		// Log the request
		Logger(start, r)

		inner.ServeHTTP(w, r)
	})
}
