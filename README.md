# Flox workshop with examples
 
A tiny HTTP API that serves quotes - used to demonstrate Flox environments.

## Lab 0: Explore our example app

This is a simple Go application that serves quotes over HTTP.

Take a look at the code:

```bash
cat main.go
```

It loads quotes from a JSON file and serves them on port 3000.

Let's try to run it:

```bash
go run main.go quotes.json
# command not found: go
```

Go is not installed. We need a development environment.

## Lab 1: Your first Flox environment

Let's fix that with Flox:

```bash
# Create a minimal Flox environment
flox init --bare

# Install Go
flox install go

# Activate the environment
flox activate

# Now it works!
go run main.go quotes.json
```

In another terminal, test it:

```bash
curl http://localhost:3000/ | jq
curl http://localhost:3000/quotes | jq
curl http://localhost:3000/quotes/0 | jq
```

## Lab 2: Running a database

Our app can also load quotes from Redis. Let's set that up as a service.

Install Redis:

```bash
flox install redis
```

Configure Redis as a service by editing the manifest:

```bash
flox edit
```

Add the following to your `manifest.toml`:

```toml
[services.redis]
command = "redis-server --port $REDISPORT"

[vars]
REDISPORT = "6379"

[profile]
common = """
  alias load_quotes='redis-cli -p $REDISPORT SET quotesjson "$(cat quotes.json)"'
"""
```

Now start the services and activate:

```bash
flox activate --start-services
```

Load quotes into Redis:

```bash
load_quotes
```

Run the app with Redis:

```bash
go run main.go redis
```

Test it:

```bash
curl http://localhost:3000/quotes | jq
```

## Lab 3: Reusing environments (composition)

Instead of configuring everything ourselves, we can reuse environments from
FloxHub. Let's replace our manifest with includes.

Edit the manifest:

```bash
flox edit
```

Replace the contents with:

```toml
version = 1

[install]

[vars]

[include]
environments = ["flox/go", "flox/redis"]
```

That's it! The `flox/go` environment provides Go, and `flox/redis` provides
Redis pre-configured as a service.

Reactivate to pick up the changes:

```bash
flox activate --start-services
```

Load quotes and run:

```bash
redis-cli -p $REDISPORT SET quotesjson "$(cat quotes.json)"
go run main.go redis
```

Environment composition lets you build on top of curated, reusable building
blocks instead of starting from scratch every time.

## Lab 4: Prepare for production (build & publish)

Let's package our app for distribution.

Edit the manifest:

```bash
flox edit
```

Add a build section:

```toml
[build.quotes-app]
command = """
  mkdir -p $out/bin $out/share
  cp quotes.json $out/share/
  go build -trimpath -o $out/bin/quotes-app main.go
"""
```

Build the package:

```bash
flox build
```

Test the built binary:

```bash
./result-quotes-app/bin/quotes-app ./result-quotes-app/share/quotes.json
```

Now publish it to FloxHub:

```bash
flox publish
```

Others can now use your published package in their environments.

Your app is now available as a reusable package on FloxHub!

---

## Reference

### Usage

```bash
# Load from a JSON file
go run main.go quotes.json

# Load from Redis
go run main.go redis
```

### Redis Setup

When using Redis as source, populate the data first:

```bash
redis-cli SET quotesjson "$(cat quotes.json)"
```

Configure Redis port via environment variable (default: 6379):

```bash
REDISPORT=6379 go run main.go redis
```

### Testing

```bash
go test -v ./...
```
