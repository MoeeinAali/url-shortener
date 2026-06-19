# --- build stage ----------------------------------------------------------
FROM docker.arvancloud.ir/golang:1.25-alpine AS build

WORKDIR /src

# Use an accessible Go module proxy (proxy.golang.org is often unreachable).
ENV GOPROXY=https://goproxy.cn,direct

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Build all three binaries.
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api       ./cmd/api && \
    CGO_ENABLED=0 go build -o /out/relay     ./cmd/relay && \
    CGO_ENABLED=0 go build -o /out/projector ./cmd/projector

# --- runtime stage --------------------------------------------------------
FROM docker.arvancloud.ir/alpine:3.20

RUN adduser -D -u 10001 appuser
WORKDIR /app

COPY --from=build /out/api       /app/api
COPY --from=build /out/relay     /app/relay
COPY --from=build /out/projector /app/projector

USER appuser

# Default entrypoint is the API; the compose file overrides the command for the
# relay and projector services.
CMD ["/app/api"]
