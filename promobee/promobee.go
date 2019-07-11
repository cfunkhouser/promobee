package promobee

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/cfunkhouser/egobee"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type thermostatMetrics struct {
	tempMetric *prometheus.GaugeVec
}

func newThermostatMetrics(t *egobee.Thermostat) *thermostatMetrics {
	return &thermostatMetrics{
		tempMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "temperature_fahrenheit",
				Help: "Temperature in Fahrenheit as reported by an Ecobee sensor.",
			},
			[]string{"location"}),
	}
}

var thermostatSelection = &egobee.Selection{
	SelectionType:   egobee.SelectionTypeRegistered,
	IncludeDevice:   true,
	IncludeRuntime:  true,
	IncludeSensors:  true,
	IncludeSettings: true,
}

// Accumulator of Ecobee information for reexport.
type Accumulator struct {
	client *egobee.Client
	done   chan<- bool

	mu          sync.RWMutex // protects following members
	thermostats map[string]*thermostatMetrics
}

func (a *Accumulator) metricsForThermostat(thermostat *egobee.Thermostat) *thermostatMetrics {
	a.mu.RLock()
	t, ok := a.thermostats[thermostat.Identifier]
	a.mu.RUnlock()

	if !ok {
		t = newThermostatMetrics(thermostat)
		a.mu.Lock()
		a.thermostats[thermostat.Identifier] = t
		a.mu.Unlock()
	}

	return t
}

func (a *Accumulator) poll() error {
	thermostats, err := a.client.Thermostats(thermostatSelection)
	if err != nil {
		return err // This error is unrecoverable.
	}
	if len(thermostats) < 1 {
		log.Printf("Payload contained no thermostats.")
		// Not technically an error. Just inconvenient.
		return nil
	}
	for _, thermostat := range thermostats {
		if len(thermostat.RemoteSensors) < 1 {
			log.Printf("Thermostat has no sensors.")
			continue
		}
		m := a.metricsForThermostat(thermostat)
		for _, sensor := range thermostat.RemoteSensors {
			t, err := sensor.Temperature()
			if err != nil {
				// We may still be able to get useful information from the payload,
				// so skip this error.
				log.Printf("Error getting temperature from %q: %v", sensor.Name, err)
				continue
			}
			m.tempMetric.With(prometheus.Labels{"location": sensor.Name}).Set(t)
		}
	}
	return nil
}

// ServeThermostatsList is a http.HandlerFunc which serves the list of known
// Thermostat identifers.
func (a *Accumulator) ServeThermostatsList(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	ids := make([]string, 0)
	a.mu.RLock()
	for id := range a.thermostats {
		ids = append(ids, id)
	}
	a.mu.RUnlock()

	sort.Strings(ids) // consistency!
	for _, id := range ids {
		fmt.Fprintf(w, "%v\n", id)
	}
}

// ServeThermostat is a http.HandlerFunc which serves the
func (a *Accumulator) ServeThermostat(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not Found")
		return
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	t, ok := a.thermostats[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not Found")
		return
	}

	registry := prometheus.NewRegistry()
	if err := registry.Register(t.tempMetric); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}
	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, req)
}

// Stop polling the Ecobee API.
func (a *Accumulator) Stop() {
	a.done <- true
}

// The Ecobee API docs recommend polling no more frequently than 3 minutes.
var defaultPollInterval = time.Minute * 3

// Opts for the Accumulator.
type Opts struct {
	PollInterval time.Duration
}

func (o *Opts) pollInterval() time.Duration {
	if o == nil || o.PollInterval == 0 {
		return defaultPollInterval
	}
	return o.PollInterval
}

// New Accumulator.
func New(c *egobee.Client, o *Opts) *Accumulator {
	done := make(chan bool)
	a := &Accumulator{
		client:      c,
		done:        done,
		thermostats: make(map[string]*thermostatMetrics),
	}

	go func(a *Accumulator, done <-chan bool) {
		ticker := time.NewTicker(o.pollInterval())
		if err := a.poll(); err != nil {
			log.Printf("error polling: %v", err)
		}
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := a.poll(); err != nil {
					log.Printf("error polling: %v", err)
				}
			}
		}
	}(a, done)

	return a
}
