FROM golang:1.26.3-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 go build -buildvcs=false -o /kronk-example .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	libgomp1 \
	libstdc++6 \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /kronk-example /app/kronk-example

ENV KRONK_BASE_PATH=/data/kronk
ENV PORT=8080

RUN mkdir -p /data/kronk

EXPOSE 8080

VOLUME ["/data/kronk"]

CMD ["/app/kronk-example"]
