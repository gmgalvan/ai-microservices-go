# Hello API Kronk

Simple Go API using Echo + Kronk.

Endpoints:

- `GET /`
- `GET /api/hello`
- `GET /api/bye`

The model used is:

- `unsloth/Qwen3-0.6B-Q8_0`

## Structure

- `main.go`: main HTTP API
- `Dockerfile`: standard image, downloads the model during `docker run`
- `Dockerfile.model`: image that preloads `llama.cpp` and the model during `docker build`
- `Dockerfile.optimized`: final image that copies the preloaded model and does not download during `docker run`
- `cmd/preload/main.go`: helper used to preload the model during image build

## Run Locally

From this folder:

```bash
cd hello-api-kronk
go run main.go
```

Test:

```bash
curl http://localhost:8080/
curl http://localhost:8080/api/hello
curl http://localhost:8080/api/bye
```

By default, the local run uses:

```bash
$HOME/.kronk
```

If you want to use an explicit local folder:

```bash
cd hello-api-kronk
export KRONK_BASE_PATH=$(pwd)/kronk-data
go run main.go
```

## Standard Docker

Build:

```bash
cd hello-api-kronk
docker build -t hello-api-kronk .
```

Run with a persistent volume:

```bash
docker run --rm -p 8080:8080 -v kronk_data:/data/kronk hello-api-kronk
```

Behavior:

- `docker build`: compiles the app and does not download the model
- first `docker run`: downloads `llama.cpp` and the model
- later `docker run` calls: reuse the volume and show `already installed`

## Docker With Preloaded Model

Build the model image:

```bash
cd hello-api-kronk
docker build -f Dockerfile.model -t hello-api-kronk-model .
```

Build the optimized image:

```bash
docker build -f Dockerfile.optimized -t hello-api-kronk-optimized .
```

Run:

```bash
docker run --rm -p 8080:8080 hello-api-kronk-optimized
```

Behavior:

- `Dockerfile.model`: downloads `llama.cpp` and the model during `docker build`
- `Dockerfile.optimized`: copies `/data/kronk` from the model image
- `docker run`: does not download anything and should show `already installed`

Important:

- do not mount an empty volume over `/data/kronk` if you want to use the model already bundled into the optimized image
- if you mount a volume on `/data/kronk`, that volume replaces what is included in the image

## Useful Commands

Show image sizes:

```bash
docker images | grep hello-api-kronk
```

Show the volume contents:

```bash
docker run --rm -v kronk_data:/data/kronk debian:bookworm-slim ls -R /data/kronk
```

## Quick Summary

Use `Dockerfile` when:

- you want a lighter image
- you do not mind downloading on the first startup
- you want persistence through a volume

Use `Dockerfile.model` + `Dockerfile.optimized` when:

- you want the model bundled into the image
- you want to avoid downloads during `docker run`
- you are fine with a much larger image
