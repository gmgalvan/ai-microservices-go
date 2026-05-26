ARG MODEL_IMAGE=lambda-api-kronk-model:latest

FROM ${MODEL_IMAGE} AS modeldata

FROM golang:1.26.3-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 go build -buildvcs=false -o /lambda-api-kronk .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
	ca-certificates \
	libgomp1 \
	libstdc++6 \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /lambda-api-kronk /app/lambda-api-kronk
COPY --from=modeldata /opt/kronk /opt/kronk

ENV KRONK_BASE_PATH=/opt/kronk

CMD ["/app/lambda-api-kronk"]
