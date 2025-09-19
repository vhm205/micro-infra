# ---- builder ----
FROM golang:1.23.10-alpine AS builder
# nếu muốn khác version: thay golang:1.21-alpine -> golang:<version>-alpine

RUN apk add --no-cache git build-base

WORKDIR /src

# Copy go.mod first to leverage Docker layer cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build args (có thể override từ docker build --build-arg ...)
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG CGO_ENABLED=0
ARG BUILD_TAGS=""
ARG VERSION=dev
ARG LDFLAGS="-s -w"

ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# Example: put version into binary: -X main.version=${VERSION}
RUN go build -trimpath -o /worker main.go

# ---- final (distroless static) ----
FROM gcr.io/distroless/static-debian11 AS runtime
# Nếu muốn dễ debug locally, thay bằng: alpine:3.18 và copy libc, hoặc use scratch for pure static.

# Create non-root user (numeric uid/gid recommended for distroless)
# Distroless images don't include useradd; we can set UID when running.
# We'll create a directory for runtime.
WORKDIR /app
COPY --from=builder /worker /app/worker

# Expose health port if worker listens for metrics/health
# (optional) EXPOSE 8080

# Use a non-root user at runtime via USER <uid>:<gid>.
# Choose a high uid like 10001 to avoid collisions.
USER 10001

ENTRYPOINT ["/app/worker"]

