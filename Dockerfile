FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

ARG TARGETOS=linux
ARG TARGETARCH
ARG TARGETVARIANT
ARG APP_VERSION=dev
ARG VCS_REF=unknown
ARG DATE_CREATED=unknown

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -trimpath -buildvcs=false \
    -ldflags="-s -w -X github.com/hatamiarash7/ipset-exporter/internal/buildinfo.Version=${APP_VERSION} -X github.com/hatamiarash7/ipset-exporter/internal/buildinfo.Revision=${VCS_REF} -X github.com/hatamiarash7/ipset-exporter/internal/buildinfo.BuildDate=${DATE_CREATED}" \
    -o /out/ipset-exporter ./cmd/ipset-exporter

FROM scratch

ARG APP_VERSION=dev
ARG VCS_REF=unknown
ARG DATE_CREATED=unknown

LABEL org.opencontainers.image.created="${DATE_CREATED}" \
    org.opencontainers.image.version="${APP_VERSION}" \
    org.opencontainers.image.revision="${VCS_REF}" \
    org.opencontainers.image.authors="Arash Hatami <info@arash-hatami.ir>" \
    org.opencontainers.image.vendor="hatamiarash7" \
    org.opencontainers.image.title="ipset-exporter" \
    org.opencontainers.image.description="Prometheus exporter for Linux ipset entry counts" \
    org.opencontainers.image.source="https://github.com/hatamiarash7/ipset-exporter" \
    org.opencontainers.image.url="https://github.com/hatamiarash7/ipset-exporter" \
    org.opencontainers.image.documentation="https://github.com/hatamiarash7/ipset-exporter#readme" \
    org.opencontainers.image.licenses="MIT"

WORKDIR /app
ENV CONFIG_FILE=/etc/ipset-exporter/config.yml
COPY --from=builder /out/ipset-exporter /app/ipset-exporter
COPY config.yml.example /etc/ipset-exporter/config.yml

# Root plus a narrowly scoped runtime NET_ADMIN capability is needed for netlink.
# No shell or package manager is included in the runtime image.
# hadolint ignore=DL3002
USER 0:0
EXPOSE 4613
ENTRYPOINT ["/app/ipset-exporter"]
