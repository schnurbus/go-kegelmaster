# syntax=docker/dockerfile:1.7
# Multi-stage build: frontend -> backend (with embedded frontend) -> runtime.
# Build with version: docker build --build-arg VERSION=v1.2.3 ...

ARG VERSION
FROM node:20-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM golang:1.25-alpine AS backend
ARG VERSION
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend /app/dist ./internal/server/web
RUN CGO_ENABLED=0 GOOS=linux go build -tags embed -ldflags "-X main.Version=${VERSION}" -o /out/kegelmaster ./cmd/api

FROM alpine:3.20
RUN adduser -D -g '' appuser
USER appuser
COPY --from=backend /out/kegelmaster /usr/local/bin/kegelmaster
EXPOSE 8080
ENV APP_ENV=production \
    BACKEND_PORT=8080 \
    JWT_SECRET=change-me \
    JWT_TTL_MINUTES=1440 \
    CORS_ALLOW_ORIGINS=*
CMD ["kegelmaster"]
