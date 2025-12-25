FROM golang:1.25 AS go

RUN CGO_ENABLED=0 go install golang.org/x/tools/gopls@latest

FROM ghcr.io/oraios/serena:latest

# install deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl git && \
    rm -rf /var/lib/apt/lists/*

# copy Go toolchain
COPY --from=go /usr/local/go /usr/local/go

# copy gopls
COPY --from=go /go/bin/gopls /usr/local/bin/gopls

ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH=$GOROOT/bin:$GOPATH/bin:$PATH

