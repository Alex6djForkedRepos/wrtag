# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine3.20 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY . .
RUN  \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/ ./cmd/...

FROM alpine:3.20
LABEL org.opencontainers.image.source=https://github.com/sentriz/mrtag
RUN apk add -U --no-cache \
    rsgain
COPY --from=builder /out/* /usr/local/bin/
CMD ["mrtagweb"]
