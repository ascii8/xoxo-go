# About

Contains a Go implementation of Tic-Tac-Toe (aka "XOXO"), written for the
Nakama game server. Includes pure Go implementations of a Tic-Tac-Toe Nakama
module, end-to-end unit tests for the Nakama module, and a Ebitengine client
that works with the Nakama module.

Showcases the end-to-end use of the
[`github.com/ascii8/nakama-go`](https://github.com/ascii8/nakama-go) and
[`github.com/ascii8/nktest`](https://github.com/ascii8/nktest) packages.

[![Tests](https://github.com/ascii8/xoxo-go/workflows/Test/badge.svg)](https://github.com/ascii8/xoxo-go/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/github.com/ascii8/xoxo-go)](https://goreportcard.com/report/github.com/ascii8/xoxo-go)
[![Reference](https://pkg.go.dev/badge/github.com/ascii8/xoxo-go.svg )](https://pkg.go.dev/github.com/ascii8/xoxo-go)

## Overview

An overview of the primary directories in this repository:

* [xoxo](/xoxo) - Tic-Tac-Toe game logic and client in Go
* [nkxoxo](/nkxoxo) - a Tic-Tac-Toe Nakama module
* [ebxoxo](/ebxoxo) - a Ebitengine game client for Tic-Tac-Toe
* [fynexoxo](/fynexoxo) - a Fyne UI game client for Tic-Tac-Toe
* [gioxoxo](/gioxoxo) - a Gio UI game client for Tic-Tac-Toe

#### Command/Module entry points

* [cmd/nkxoxo](/cmd/nkxoxo) - the Nakama server module entry point
* [cmd/nkclient](/cmd/nkclient) - the testing client
* [cmd/ebclient](/cmd/ebclient) - the Ebitengine client entry point
* [cmd/fyneclient](/cmd/fyneclient) - the Fyne UI client entry point
* [cmd/gioclient](/cmd/gioclient) - the Gio UI client entry point

## Running the Unit Tests

Checkout the code and run the tests using `go test` from the repository root:

```sh
# get the code
$ git clone https://github.com/ascii8/xoxo-go.git && cd xoxo-go

# build/run the Nakama module with Nakama server, and run the unit tests
$ DEBUG=1 go test -v
```

## Running the Module for use with Clients

Run the module using `go test` from the repository root:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build/run the Nakama module with Nakama server
$ DEBUG=1 KEEP=2h go test -v -timeout=2h -run TestKeep
```

## Using the Ebitengine client (Desktop)

Build and run the Ebitengine client, as a Desktop client:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build/run the Ebitengine client
$ go build ./cmd/ebclient && ./ebclient
```

## Using the Ebitengine client (WASM)

Build and run the Ebitengine client as a WASM module in a Web Browser:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build wasm and run local webserver
$ go run github.com/hajimehoshi/wasmserve@latest ./cmd/ebclient
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080) in a browser.

See: [Ebitengine WASM documentation](https://ebitengine.org/en/documents/webassembly.html)

## Using the Fyne client (Desktop)

Build and run the Fyne client, as a Desktop client:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build/run the Fyne client
$ go build ./cmd/fyneclient && ./fyneclient
```

## Using the Fyne client (WASM)

Build and run the Fyne client, as a WASM client:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build wasm and run local webserver
$ go run github.com/hajimehoshi/wasmserve@latest ./cmd/fyneclient
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080) in a browser.

## Using the Gio client (Desktop)

Build and run the Gio client, as a Desktop client:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build/run the Gio client
$ go build ./cmd/gioclient && ./gioclient
```

## Using the Gio client (WASM)

Build and run the Gio client, as a WASM client:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build wasm and run local wgioserver
$ go run github.com/hajimehoshi/wasmserve@latest ./cmd/gioclient
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080) in a browser.

## Using the Defold client

1. Grab Defold client code, and configure:

```sh
# get the Defold client
$ git clone https://github.com/defold/game-xoxo-nakama-client.git && cd game-xoxo-nakama-client

# change game.project settings
$ perl -pi -e 's/host =.*/host = 127.0.0.1/' game.project
$ perl -pi -e 's/port =.*/port = 7352/' game.project
$ perl -pi -e 's/server_key =.*/server_key = xoxo-go_server/' game.project
```

2. Build and run the Defold client:

```sh
# change to path
$ cd /path/to/game-xoxo-nakama-client

# build and fix permissions
$ java -jar /opt/Defold/bob.jar --variant=debug && chmod +x ./build/x86_64-linux/dmengine

# run defold with debugging
$ DM_SERVICE_PORT=dynamic ./build/x86_64-linux/dmengine
```

## Related Links

* [github.com/ascii8/nakama-go](https://github.com/ascii8/nakama-go) - a Nakama client for Go, with realtime WebSocket and WASM support
* [github.com/ascii8/nktest](https://github.com/ascii8/nktest) - a Nakama module testing package for Go
* [github.com/defold/game-xoxo-nakama-client](https://github.com/defold/game-xoxo-nakama-client.git) - a Nakama Tic-Tac-Toe client made with Defold
* [github.com/defold/game-xoxo-nakama-server](https://github.com/defold/game-xoxo-nakama-server.git) - a Nakama Tic-Tac-Toe server (Lua)
