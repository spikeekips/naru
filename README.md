# NARU

API server for SEBAK. NARU does nothing :)

## Build

```sh
$ go build github.com/spikeekips/naru/cmd/naru
```

## Test

```sh
$ go test ./... -v
```

## Deploy

Digest first,
```sh
$ ./naru digest \
    --sebak http://localhost:12345 \
    --jsonrpc http://localhost:54321/jsonrpc \
    --storage file:///tmp/db
```

Run server,
```sh
$ ./naru server \
    --bind http://localhost:23456 \
    --sebak http://localhost:12345 \
    --jsonrpc http://localhost:54321/jsonrpc \
    --storage file:///tmp/db
```

* This will connect to sebak,
  - sebak node endpoint: http://localhost:12345
  - sebak jsonrpc: http://localhost:54321
