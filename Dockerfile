# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:fe69f483d2ef3e2309ecb19d18ab01054711746bc0a31bf5fbcffccb32182f05 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build-batch

FROM debian:bullseye-slim@sha256:610b4c7ad241e66f6e2f9791e3abdf0cc107a69238ab21bf9b4695d51fd6366a
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/batch ./
CMD ["./batch"]
