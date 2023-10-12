# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:9bcf629934ea0d2544d0a6a03288ca1956e6c30d8fe36763494e260ef9b9c6e6 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:9bec46ecd98ce4bf8305840b021dda9b3e1f8494a0768c407e2b233180fa1466
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/annict2anilist ./
CMD ["./annict2anilist"]
