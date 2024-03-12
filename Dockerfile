# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:fe69f483d2ef3e2309ecb19d18ab01054711746bc0a31bf5fbcffccb32182f05 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build-batch

FROM debian:bullseye-slim@sha256:a165446a88794db4fec31e35e9441433f9552ae048fb1ed26df352d2b537cb96
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/batch ./
CMD ["./batch"]
