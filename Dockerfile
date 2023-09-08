# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:b120a1cc1cab8358044d86ed64472685cda9283c95c66db9ec32d149bba46082 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:3bc5e94a0e8329c102203c3f5f26fd67835f0c81633dd6949de0557867a87fac
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/annict2anilist ./
CMD ["./annict2anilist"]
