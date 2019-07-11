package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cfunkhouser/egobee"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/urfave/cli.v2"

	"github.com/cfunkhouser/promobee/promobee"
)

const (
	authURLTemplate = `https://api.ecobee.com/authorize?response_type=ecobeePin&scope=smartWrite&client_id=%v`
	tokenURL        = "https://api.ecobee.com/token"
)

var (
	// The Ecobee API docs recommend polling no more frequently than 3 minutes.
	pollInterval = time.Minute * 3
)

func init() {
	flag.DurationVar(&pollInterval, "poll_interval", time.Minute*3, "Interval at which to poll the Ecobee API for updates.")
}

func main() {
	app := &cli.App{
		Name:        "promobee",
		Description: "Export Ecobee details to Prometheus",
		Version:     "0.1.0",
		Authors: []*cli.Author{
			{Name: "Christian Funkhouser", Email: "christian.funkhouser@gmail.com"},
			{Name: "Sarah Funkhouser", Email: "sarah.k.funkhouser@gmail.com"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api_key",
				Aliases: []string{"app", "k"},
				Usage:   "Ecobee API Key. Required.",
				EnvVars: []string{"PROMOBEE_API_KEY"},
			},
			&cli.StringFlag{
				Name:    "store",
				Aliases: []string{"s"},
				Usage:   "Ecobee API credential token store file location. Required.",
				EnvVars: []string{"PROMOBEE_TOKEN_STORE"},
			},
			&cli.Uint64Flag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "Port on which to serve Prometheus metrics",
				EnvVars: []string{"PORT", "PROMOBEE_PORT"},
				Value:   8080,
			},
			&cli.StringFlag{
				Name:    "address",
				Aliases: []string{"a"},
				Usage:   "Address to bind for serving Prometheus metrics",
				EnvVars: []string{"PROMOBEE_ADDRESS"},
			},
			&cli.StringFlag{
				Name:    "httplog",
				Usage:   "If set to a file path, all HTTP requests and responses will be logged there.",
				EnvVars: []string{"PROMOBEE_HTTP_LOG"},
			},
		},
		Action: doServeMetrics,
		Commands: []*cli.Command{
			{
				Name:        "register",
				Usage:       "Register Promobee application with Ecobee account",
				Description: "Registers Promobee application with Ecobee account",
				Action:      doRegister,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func doServeMetrics(c *cli.Context) error {
	hostPort := fmt.Sprintf("%v:%d", c.String("address"), c.Uint64("port"))

	opts := &egobee.Options{}
	if httpLog := c.String("httplog"); httpLog != "" {
		f, err := os.OpenFile(httpLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return cli.Exit(fmt.Errorf("failed creating http log %q: %v", httpLog, err), 1)
		}
		opts.Log = true
		opts.LogTo = f
	}

	storePath := c.String("store")
	if storePath == "" {
		cli.ShowAppHelpAndExit(c, 1)
	}
	ts, err := egobee.NewPersistentTokenFromDisk(storePath)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed initializing store %q: %v", storePath, err), 1)
	}

	apiKey := c.String("api_key")
	if apiKey == "" {
		cli.ShowAppHelpAndExit(c, 1)
	}
	p := promobee.New(egobee.New(apiKey, ts, opts), nil)

	// Export the default metrics.
	http.Handle("/metrics", promhttp.Handler())

	// Export Ecobee metrics
	http.HandleFunc("/thermostats", p.ServeThermostatsList)
	http.HandleFunc("/thermostat", p.ServeThermostat)

	log.Printf("Starting on %v", hostPort)
	return http.ListenAndServe(hostPort, nil)
}

func doRegister(c *cli.Context) error {
	storePath := c.String("store")
	if storePath == "" {
		cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
	}
	apiKey := c.String("api_key")
	if apiKey == "" {
		cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
	}

	resp, err := http.Get(fmt.Sprintf(authURLTemplate, apiKey))
	if err != nil {
		return cli.Exit(fmt.Errorf("failed initializing Pin Authentication: %v", err), 1)
	}

	pac := &egobee.PinAuthenticationChallenge{}
	if err := json.NewDecoder(resp.Body).Decode(pac); err != nil {
		return cli.Exit(fmt.Errorf("failed to read Pin Authentication: %v", err), 1)
	}
	resp.Body.Close()

	fmt.Printf("Register with this PIN: %v\n", pac.Pin)
	fmt.Println("Press any key to continue when done.")

	var input string
	fmt.Scanf("%s", &input)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "grant_type=ecobeePin&code=%v&client_id=%v", pac.AuthorizationCode, apiKey)

	resp, err = http.Post(tokenURL, "application/x-www-form-urlencoded", &buf)
	if err != nil {
		return cli.Exit(fmt.Errorf("failed authenticating: %v", err), 1)
	}
	defer resp.Body.Close()

	trr := &egobee.TokenRefreshResponse{}
	if err := json.NewDecoder(resp.Body).Decode(trr); err != nil {
		return cli.Exit(fmt.Errorf("failed decoding authentication response: %v", err), 1)
	}
	if _, err = egobee.NewPersistentTokenStore(trr, storePath); err != nil {
		return cli.Exit(fmt.Errorf("failed creating persistent store: %v", err), 1)
	}
	fmt.Printf("Created persistent store at %v\n", storePath)
	return nil
}
