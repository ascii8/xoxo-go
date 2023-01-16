# About

Contains a Go implementation of Tic-Tac-Toe (aka "xoxo"), written for the
Nakama game server. Includes pure Go implementations of a Tic-Tac-Toe Nakama
module, end-to-end unit tests for the module, and a Ebitengine client that
works with the module.

Showcases the end-to-end use of the
[`github.com/ascii8/nakama-go`](https://github.com/ascii8/nakama-go) and
[`github.com/ascii8/nktest`](https://github.com/ascii8/nktest) packages.

## Overview

An overview of the directories in this repository:

* [xoxo](/xoxo) - Tic-Tac-Toe game logic and client in Go
* [nkxoxo](/nkxoxo) - a Tic-Tac-Toe Nakama module
* [xoxo-cli](/cmd/xoxo-cli) - a non-interactive command-line client for Tic-Tac-Toe (randomly selects an available cell)
* [ebxoxo](/ebxoxo) - a Ebitengine game client for Tic-Tac-Toe

#### Command/Module entry points

* [cmd/nkxoxo](/cmd/nkxoxo) - the Nakama module entry point
* [cmd/ebxoxo](/cmd/ebxoxo) - the Ebitengine client entry point

## Running the Unit Tests

Checkout the code and run the tests using `go test` from the repository root:

```sh
# get the code
$ git clone https://github.com/ascii8/xoxo-go.git && cd xoxo-go

# build/run the Nakama module with Nakama server, and run the unit tests
$ DEBUG=1 go test -v
```

## Running the Module for use with Clients

Rrun the module using `go test` from the repository root:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build/run the Nakama module with Nakama server
$ DEBUG=1 KEEP=2h go test -v -timeout=2h -run TestKeep
```

## Using the Ebitengine client

Build and run the Ebitengine client:

```sh
# change to the repository root
$ cd /path/to/xoxo-go

# build/run the Ebitengine client
$ go build ./cmd/ebxoxo && ./ebxoxo
```

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
