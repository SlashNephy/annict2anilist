# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:436969571fa091f02d34bf2b9bc8850af7de0527e5bc53c39eeda88bc01c38d3 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:f794067bee57cf99c4c2b32f022a8782ad47c89e11f935d3ca5fdc7414cc465d
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/annict2anilist ./
CMD ["./annict2anilist"]
