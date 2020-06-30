package promobee

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cfunkhouser/egobee"
)

func TestAccumulator_ServeThermostatList(t *testing.T) {
	acc := &Accumulator{
		thermostats: map[string]*thermostatMetrics{
			"id1": &thermostatMetrics{},
			"id2": &thermostatMetrics{},
			"id3": &thermostatMetrics{},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "/thermostats", nil)
	if err != nil {
		t.Fatalf("failed creating request: %v", err)
	}

	rr := httptest.NewRecorder()
	http.HandlerFunc(acc.ServeThermostatsList).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("incorrect status; got %q, want %q", rr.Code, http.StatusOK)
	}

	wantCT := "text/plain"
	if gotCT := rr.Header().Get("Content-Type"); gotCT != wantCT {
		t.Errorf("incorrect Content-Type header; got %q, want %q", gotCT, wantCT)
	}

	want := "id1\nid2\nid3\n"
	if got := rr.Body.String(); got != want {
		t.Errorf("incorrect content; got %q, want %q", got, want)
	}
}

func TestAccumulator_metricsForThermostat_doesNotExist(t *testing.T) {
	testAccumulator := &Accumulator{
		thermostats: make(map[string]*thermostatMetrics),
	}
	got := testAccumulator.metricsForThermostatIdentifier(&egobee.Thermostat{Identifier: "foo"})
	if got == nil {
		t.Errorf("Accumulator.metricsForThermostatIdentifier(...) returned nil; it should never do that.")
	}
	if inMap := testAccumulator.thermostats["foo"]; inMap != got {
		t.Errorf("Accumulator.metricsForThermostatIdentifier(...) returned a pointer that doesn't exist in the map somehow")
	}
}

func TestAccumulator_metricsForThermostat_doesExist(t *testing.T) {
	tm := &thermostatMetrics{}
	testAccumulator := &Accumulator{
		thermostats: map[string]*thermostatMetrics{
			"foo": tm,
		},
	}
	got := testAccumulator.metricsForThermostatIdentifier(&egobee.Thermostat{Identifier: "foo"})
	if got == nil {
		t.Errorf("Accumulator.metricsForThermostatIdentifier(...) returned nil; it should never do that.")
	}
	if got != tm {
		t.Errorf("Accumulator.metricsForThermostatIdentifier(...) returned a new *thermostatMetrics despite one being in the map")
	}
}

func TestOptsPollInterval(t *testing.T) {
	for _, tt := range []struct {
		name string
		opts *Opts
		want time.Duration
	}{
		{name: "nil receiver", want: time.Minute * 3},
		{name: "zero-value interval", want: time.Minute * 3, opts: &Opts{}},
		{name: "valid non-default", want: time.Hour, opts: &Opts{PollInterval: time.Hour}},
	} {
		if got := tt.opts.pollInterval(); got != tt.want {
			t.Errorf("%v: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
