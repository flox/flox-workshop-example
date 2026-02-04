# Quotes App (Go + Redis)

A tiny HTTP API that serves quotes loaded from Redis.

## Overview

Loads a JSON array from Redis key quotesjson at startup.

Caches the array in memory and serves it over HTTP.

## Endpoints:

GET /quotes — return all quotes.

GET /quotes/{index} — return a single quote by zero-based index.

## Getting started

Start the app for local development:
`go run main.go`

Populate the DB for local development:
`redis-cli SET quotesjson "$(cat quotes.json)"`

## Build

```
mkdir -p $out/{lib,bin}
cp -pr quotes.json $out/lib
go mod tidy
go build -trimpath -o $out/bin/quotes-app-go main.go
chmod +x $out/bin/quotes-app-go
```
