FROM golang:1.13 AS builder
LABEL maintainer="Christian Funkhouser <christian@funkhouse.rs>"

COPY . ./promobee/
RUN go build -mod=vendor -o /promobee ./promobee/main.go

FROM golang:1.13
COPY --from=builder /promobee .
EXPOSE 8080
VOLUME ["/var/run/promobee"]
ENTRYPOINT [ "./promobee", "--store", "/var/run/promobee/promobee.store", "--api_key" ]