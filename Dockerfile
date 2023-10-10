# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:9bc6dcb86d0b13c6ecc41284c4ca4c940c7be322d18d9ab652c0f1af11ac9327 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:c618be84fc82aa8ba203abbb07218410b0f5b3c7cb6b4e7248fda7785d4f9946
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/annict2anilist ./
CMD ["./annict2anilist"]
