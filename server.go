package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/ubuntu-core/identity-vault/service"
)

var settingsFile string
var config service.ConfigSettings

func parseArgs() {
	flag.StringVar(&settingsFile, "config", "./settings.yaml", "Path to the config file")
	flag.Parse()
}

func readConfig() {
	source, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		log.Fatalf("Error opening the config file: %v", err)
	}

	err = yaml.Unmarshal(source, &config)
	if err != nil {
		log.Fatalf("Error parsing the config file: %v", err)
	}
}

func main() {
	// Parse the command line arguments
	parseArgs()
	readConfig()

	// Start the web service router
	router := mux.NewRouter()

	router.Handle("/1.0/version", service.Middleware(http.HandlerFunc(service.VersionHandler), &config)).Methods("GET")
	router.Handle("/1.0/sign", service.Middleware(http.HandlerFunc(service.SignHandler), &config)).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
