FROM golang:1.26.3-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY models/preload ./models/preload

RUN CGO_ENABLED=0 go build -buildvcs=false -o /preload ./models/preload

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	libgomp1 \
	libstdc++6 \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /preload /app/preload

ENV KRONK_BASE_PATH=/opt/kronk

RUN mkdir -p /opt/kronk
RUN /app/preload
