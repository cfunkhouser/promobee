# promobee

A Prometheus exporter for ecobee data written in the `Go` programming language.

`promobee` exports a list of known thermostat identifiers at `/thermostats`. The
list of identifiers can be used for target discovery purposes.

Metrics for a given thermostat are retrieved from
`/thermostat?id=$THERMOSTAT_ID`.

## Usage

You will need an API key. Read the [Reference API
App](https://www.ecobee.com/home/developer/api/sample-apps/reference-api-app.shtml)
documentation from Ecobee to understand how to get one of these.

Once you have an API key, `promobee` requires you to add the application to your
account, and that you have configured an `egobee` Token Store. Use the
`ego-register` command from that package to initialize one:

```console
$ ego-register \
    --app $ECOBEE_API_KEY \
    --store /path/to/store
Register with this PIN: abc9
Press any key to continue when done.
```

Once you have a code, go to [the Ecobee website](https://www.ecobee.com/), log
in, navigate to _My Apps_ and click _Add Application._ When prompted, enter the
code from above and click _Validate._

Now, you can run `promobee`:

```console
$ promobee \
    --app $ECOBEE_API_KEY \
    --store /path/to/store
2019/07/10 12:04:10 Starting on :8080
```

If anything happens to the token store, you will need to re-add the application.

Once `promobee` is configured, you can point Prometheus at it with a
configuration like:

```yaml
scrape_configs:
  - job_name: "promobee"
    metrics_path: /thermostat
    # The Ecobee API recommends not polling their API more than once every 3
    # minutes, which Promobee respects. Poll twice that often to help reduce
    # chances of missing an interesting point. Polling the metric endpoint does
    # not cause an API request.
    scrape_interval: 90s
    static_configs:
        - targets:
            # These are the themostat IDs as exported by the /thermostats
            # endpoint.
            - 123456789098
            - 123456789099
    relabel_configs:
            - source_labels: [__address__]
              target_label: __param_id
            - source_labels: [__param_id]
              target_label: thermostat
            - target_label: __address__
              # Replace this host:port with the location of promobee.
              replacement: 10.42.18.11:8080
```
