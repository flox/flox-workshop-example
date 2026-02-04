# Quotes App (Go)

A tiny HTTP API that serves quotes from a JSON file or Redis.

## Overview

Loads quotes lazily on first request from either a local JSON file or Redis.
Caches in memory and serves over HTTP.

## Endpoints

- `GET /quotes` — return all quotes
- `GET /quotes/{index}` — return a single quote by zero-based index

## Usage

```bash
# Load from a JSON file
go run main.go quotes.json

# Load from Redis
go run main.go redis
```

An argument is required. Running without one shows usage help:

```
Usage: quotes-app <quotes.json | redis>
  quotes.json  - path to a JSON file containing quotes
  redis        - load quotes from Redis
```

## Redis Setup

When using Redis as source, populate the data first:

```bash
redis-cli SET quotesjson "$(cat quotes.json)"
```

Configure Redis port via environment variable (default: 6379):

```bash
REDISPORT=6379 go run main.go redis
```

## Testing

```bash
go test -v ./...
```

## Build

```bash
mkdir -p $out/{lib,bin}
cp -pr quotes.json $out/lib
go mod tidy
go build -trimpath -o $out/bin/quotes-app-go main.go
chmod +x $out/bin/quotes-app-go
```
