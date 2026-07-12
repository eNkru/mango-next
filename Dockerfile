# syntax=docker/dockerfile:1
# Multi-arch Go build for Mango.
# Build: docker buildx build --platform linux/amd64,linux/arm64 -t mango .
# Or:    docker build -t mango -f go/Dockerfile go/

FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src
COPY go/ ./

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mango ./cmd/mango/

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /mango /mango

EXPOSE 9000
ENTRYPOINT ["/mango"]
