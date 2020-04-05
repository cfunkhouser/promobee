FROM golang:1.13 AS builder
LABEL maintainer="Christian Funkhouser <christian@funkhouse.rs>"

COPY . ./build/promobee/
RUN cd ./build/promobee && go build -mod=vendor -o /promobee .

# Copy from builder image to keep the size down. Resulting image should only
# contain the promobee binary itself.
FROM golang:1.13
COPY --from=builder /promobee .
EXPOSE 8080
VOLUME ["/var/run/promobee"]
ENTRYPOINT [ "./promobee", "--store", "/var/run/promobee/promobee.store", "--api_key" ]