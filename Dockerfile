# syntax=docker/dockerfile:1

ARG GO_VERSION

# ---- Build ----------------------------------------------------------------
FROM golang:${GO_VERSION} AS builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG REVISION=dev

# upx (build stage only) compresses the final binary to shrink the image.
RUN apt-get update \
    && apt-get install -y --no-install-recommends upx-ucl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${REVISION}" \
    -o echo ./cmd/echo

RUN upx --best --lzma echo

# ---- Runtime --------------------------------------------------------------
# scratch: echo is a static (CGO_ENABLED=0) binary that makes no outbound TLS
# calls and reads no user database, so it needs neither CA certs nor /etc/passwd.
# The chart's podSecurityContext runs it as non-root (uid/gid 65532).
FROM scratch
COPY --from=builder /workspace/echo /echo
ENTRYPOINT ["/echo"]
