FROM golang:1.22 as build
WORKDIR /app/src

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Leverage build cache by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
RUN go mod verify

# Copy the rest of the source code
COPY . .

RUN --mount=type=cache,target="/root/.cache/go-build" go build -o ../user-service ./cmd/user-service

FROM alpine as app

# Add curl for health checks
RUN apk --update --no-cache add curl

SHELL ["/bin/sh", "-l", "-c"]

# Create a non-root user
ARG user=app
ARG group=docker
RUN addgroup $group
RUN adduser \
    --disabled-password \
    --gecos "" \
    --ingroup $group \
    $user
USER $user
WORKDIR /usr/local/user-service

# Copy the binary and set executable permissions
COPY --chown=$user:$group --from=build /app/user-service /usr/local/bin/user-service
RUN chmod +x /usr/local/bin/user-service

# Add a health check
HEALTHCHECK --interval=5s --timeout=3s --retries=3 CMD curl --fail http://localhost:80/healthz || exit 1
EXPOSE 80

CMD ["/usr/local/bin/user-service"]