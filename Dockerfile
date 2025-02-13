FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app/
ADD . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o ipset-exporter cmd/ipset-exporter/main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

ARG DATE_CREATED
ARG APP_VERSION
ENV APP_VERSION=$APP_VERSION

LABEL org.opencontainers.image.created=$DATE_CREATED
LABEL org.opencontainers.version="$APP_VERSION"
LABEL org.opencontainers.image.authors="Arash Hatami <info@arash-hatami.ir>"
LABEL org.opencontainers.image.vendor="hatamiarash7"
LABEL org.opencontainers.image.title="ipset-exporter"
LABEL org.opencontainers.image.description="It's a simple Prometheus exporter for ipset"
LABEL org.opencontainers.image.source="https://github.com/hatamiarash7/ipset-exporter"
LABEL org.opencontainers.image.url="https://github.com/hatamiarash7/ipset-exporter"
LABEL org.opencontainers.image.documentation="https://github.com/hatamiarash7/ipset-exporter"

WORKDIR /app/

COPY --from=builder /app/ipset-exporter /app/ipset-exporter

ENTRYPOINT ["/app/ipset-exporter"]
