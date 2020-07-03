package promobee

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cfunkhouser/egobee"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type thermostatMetrics struct {
	tempMetric      *prometheus.GaugeVec
	hvacModeMetric  *prometheus.GaugeVec
	holdTempMetric  *prometheus.GaugeVec
	hvacInOperation *prometheus.GaugeVec
	humidityMetric  *prometheus.GaugeVec
	occupancyMetric *prometheus.GaugeVec
}

func newThermostatMetrics() *thermostatMetrics {
	return &thermostatMetrics{
		tempMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "temperature_fahrenheit",
				Help: "Temperature in Fahrenheit as reported by an Ecobee sensor.",
			},
			[]string{"location"}),
		holdTempMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hold_temperature_fahrenheit",
				Help: "Hold temperatures in Fahrenheit as reported by an Ecobee Thermostat",
			},
			[]string{"type"},
		),
		hvacModeMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hvac",
				Help: "HVAC mode as reported by an Ecobee thermostat.",
			},
			[]string{"mode"},
		),
		hvacInOperation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "hvac_in_operation",
				Help: "Running HVAC equipment is emitted with a '1' metric",
			},
			[]string{"equipment"},
		),
		humidityMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "humidity",
				Help: "Humidity as reported by an Ecobee sensor.",
			},
			[]string{"location"}),

		occupancyMetric: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "occupancy",
				Help: "Occupancy as reported by an Ecobee sensor.",
			},
			[]string{"location"}),
	}
}

var thermostatSelection = &egobee.Selection{
	SelectionType:   egobee.SelectionTypeRegistered,
	IncludeDevice:   true,
	IncludeEvents:   true,
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

func (a *Accumulator) metricsForThermostatIdentifier(identifier *string) *thermostatMetrics {
	a.mu.RLock()
	t, ok := a.thermostats[*identifier]
	a.mu.RUnlock()

	if !ok {
		t = newThermostatMetrics()
		a.mu.Lock()
		a.thermostats[*identifier] = t
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
		m := a.metricsForThermostatIdentifier(&thermostat.Identifier)

		m.holdTempMetric.Reset()

		if thermostat.Settings.HVACMode != "off" {
			for _, event := range thermostat.Events {
				if event.Running && event.Type == "hold" {
					if !event.IsCoolOff && thermostat.Settings.HVACMode != "heat" {
						m.holdTempMetric.WithLabelValues("cool").Set(float64(event.CoolHoldTemp) / 10)
					}
					if !event.IsHeatOff && thermostat.Settings.HVACMode != "cool" {
						m.holdTempMetric.WithLabelValues("heat").Set(float64(event.HeatHoldTemp) / 10)
					}
				}
			}
		}

		m.hvacModeMetric.Reset()
		m.hvacModeMetric.WithLabelValues(thermostat.Settings.HVACMode).Set(1)

		for _, sensor := range thermostat.RemoteSensors {
			h, err := sensor.Humidity()
			// Only handle the successful case; if the sensor doesn't have humidity, that isn't fatal
			if err == nil {
				m.humidityMetric.With(prometheus.Labels{"location": sensor.Name}).Set(float64(h))
			}

			o, err := sensor.Occupancy()
			// Only handle the successful case; if the sensor doesn't have occupancy, that isn't fatal
			if err == nil {
				v := 0.0
				if o {
					v = 1.0
				}
				m.occupancyMetric.With(prometheus.Labels{"location": sensor.Name}).Set(v)
			}

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

	statSummary, err := a.client.ThermostatSummary()
	if err != nil {
		return err
	}

	for _, status := range statSummary.StatusList {
		d := strings.Split(status, ":")
		if len(d) != 2 {
			log.Printf("Thermostat status '%s' did not have two fields", status)
			continue
		}
		m := a.metricsForThermostatIdentifier(&d[0])
		m.hvacInOperation.Reset()
		if d[1] != "" {
			for _, unit := range strings.Split(d[1], ",") {
				m.hvacInOperation.WithLabelValues(unit).Set(1)
			}
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
	metrics := []*prometheus.GaugeVec{t.tempMetric, t.occupancyMetric, t.humidityMetric, t.holdTempMetric, t.hvacInOperation, t.hvacModeMetric}
	for _, m := range metrics {
		if err := registry.Register(m); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Server Error")
			return
		}
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
