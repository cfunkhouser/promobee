# promobee

A Prometheus exporter for ecobee data written in the `Go` programming language.

## Export format

`promobee` exports a list of known thermostat identifiers at `/thermostat` if no `id` query parameter is provided. If an identifier is provided, metrics are provided for collection by Prometheus.

The list of identifiers can be used for target discovery purposes.