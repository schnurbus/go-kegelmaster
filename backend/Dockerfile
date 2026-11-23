# syntax=docker/dockerfile:1.29

FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/api

FROM alpine:3.24
WORKDIR /app
RUN adduser -D -g '' appuser
USER appuser
COPY --from=builder /out/server /usr/local/bin/server
EXPOSE 8080
ENV APP_ENV=production \
    BACKEND_PORT=8080 \
    JWT_SECRET=change-me \
    JWT_TTL_MINUTES=1440 \
    CORS_ALLOW_ORIGINS=*
CMD ["server"]

