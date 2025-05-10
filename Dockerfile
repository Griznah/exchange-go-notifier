# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o exchange-go-notifier main.go

FROM alpine:latest
WORKDIR /app
# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
COPY --from=build /app/exchange-go-notifier ./
COPY --from=build /app/api_state.json ./
EXPOSE 8080
ENV PORT=8080
CMD ["./exchange-go-notifier"]
