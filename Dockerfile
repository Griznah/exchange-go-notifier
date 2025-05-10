# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o exchange-go-notifier main.go

FROM alpine:3.21
WORKDIR /app
# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=build /app/exchange-go-notifier ./
COPY --from=build /app/api_state.json ./
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
ENV PORT=8080
CMD ["./exchange-go-notifier"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s \
   CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/health || exit 1