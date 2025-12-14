# syntax=docker/dockerfile:1
FROM golang:1.25.5-bookworm AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build-batch

FROM debian:bookworm-slim
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN <<EOF
  set -eux
  apt-get update
  apt-get install -yqq --no-install-recommends ca-certificates
  rm -rf /var/lib/apt/lists/*
EOF

USER app
COPY --from=build /app/batch ./
CMD ["./batch"]
