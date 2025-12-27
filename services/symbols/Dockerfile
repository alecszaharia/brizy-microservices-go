FROM golang:1.25 AS builder

ARG TARGETARCH
ARG TARGETOS
ARG VERSION=dev

COPY ../../ /src
WORKDIR /src/services/symbols

RUN mkdir -p bin/
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/services/symbols/bin /app

WORKDIR /app

EXPOSE 8000
EXPOSE 7000
VOLUME /data/conf

CMD [ "./symbol-service", "--conf", "/data/conf/config.yaml" ]
