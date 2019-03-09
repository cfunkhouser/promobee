package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cfunkhouser/egobee"
	"github.com/cfunkhouser/promobee/promobee"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr      = flag.String("address", ":8080", "The address and port on which to serve HTTP.")
	appID     = flag.String("app", "", "Ecobee Registered App ID")
	storePath = flag.String("store", "/tmp/promobee", "Persistent egobee credential store path")
	httpLog   = flag.String("http_log", "", "If set to a file path, all HTTP requests and responses will be logged there.")

	// The Ecobee API docs recommend polling no more frequently than 3 minutes.
	pollInterval = time.Minute * 3
)

func init() {
	flag.DurationVar(&pollInterval, "poll_interval", time.Minute*3, "Interval at which to poll the Ecobee API for updates.")
}

func main() {
	flag.Parse()

	if *appID == "" {
		log.Fatalf("--app is required")
	}
	if *storePath == "" {
		log.Fatalf("--store is required")
	}

	opts := &egobee.Options{}
	if *httpLog != "" {
		f, err := os.OpenFile(*httpLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file %q for append: %v", *httpLog, err)
		}
		opts.Log = true
		opts.LogTo = f
	}

	ts, err := egobee.NewPersistentTokenFromDisk(*storePath)
	if err != nil {
		log.Fatalf("Failed to initialize store %q: %v", *storePath, err)
	}

	client := egobee.New(*appID, ts, opts)
	p := promobee.New(client, nil)

	// Export the default metrics.
	http.Handle("/metrics", promhttp.Handler())

	// Export Ecobee metrics
	http.HandleFunc("/thermostats", p.ServeThermostatsList)
	http.HandleFunc("/thermostat", p.ServeThermostat)

	log.Printf("Starting on %v", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
