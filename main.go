package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/cfunkhouser/egobee"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// The Ecobee App ID Key for Promobee.
const appID = "94fl7gSO9SFmTXBKr0aSjzMwkjXAIRnZ"

var (
	addr = flag.String("address", ":8080", "The address and port on which to serve HTTP.")

	// These flags should be set with the initial access/refresh tokens. The
	// egobee library will refresh the tokens in memory during the life of the
	// process.
	// TODO(cfunkhouser): Use the persistent store when available.
	accessToken  = flag.String("access_token", "", "Initial Ecobee API Access Token")
	refreshToken = flag.String("refresh_token", "", "Initial Ecobee API Refresh Token")

	tempMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "temperature_fahrenheit",
			Help: "Temperature in Fahrenheit as reported by an Ecobee sensor.",
		},
		[]string{"location"})

	thermostatSelection = &egobee.Selection{
		SelectionType:   egobee.SelectionTypeRegistered,
		IncludeSettings: true,
		IncludeRuntime:  true,
		IncludeSensors:  true,
	}
)

func init() {
	prometheus.MustRegister(tempMetric)
}

func main() {
	log.Print("Starting...")
	flag.Parse()

	if *accessToken == "" || *refreshToken == "" {
		log.Fatalf("--access_token and --refresh_token are both required.")
	}

	http.Handle("/metrics", promhttp.Handler())

	ts := egobee.NewMemoryTokenStore(&egobee.TokenRefreshResponse{
		AccessToken:  *accessToken,
		RefreshToken: *refreshToken,
		ExpiresIn:    egobee.TokenDuration{Duration: time.Minute * 15},
	})

	client := egobee.New(appID, ts)

	populate := func() {
		thermostats, err := client.Thermostats(thermostatSelection)
		if err != nil {
			log.Printf("Hrm, no good; skipping round: %+v", err)
		}
		for _, thermostat := range thermostats {
			// TODO(cfunkhouser): Expose averages.
			if len(thermostat.RemoteSensors) > 0 {
				for _, sensor := range thermostat.RemoteSensors {
					t, err := sensor.Temperature()
					if err != nil {
						log.Printf("Error getting temperature from %q: %v", sensor.Name, err)
						continue
					}
					tempMetric.With(prometheus.Labels{"location": sensor.Name}).Set(t)
				}
			}
		}
	}

	go func() {
		// The Ecobee API docs recommend polling no more frequently than 3 minutes.
		ticker := time.NewTicker(time.Minute * 3)
		populate()
		for {
			select {
			case <-ticker.C:
				populate()
			}
		}

	}()

	log.Fatal(http.ListenAndServe(*addr, nil))
}
