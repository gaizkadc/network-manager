# gRPC example server

First, built and install the server

```
$ make build
```

Then, launch the server with:

```
$ ./bin/grpcserver --consoleLogging api
```

You can use `grpc_cli` to check which services are running

```
$ grpc_cli ls localhost:8000
grpc.reflection.v1alpha.ServerReflection
ping.Ping
```

## Available options:

```
gRPC example server for testing Go gRPC

Usage:
  grpcserver [command]

Available Commands:
  api         Launch the gRPC services
  help        Help about any command
  version     Print the version number of gRPC

Flags:
      --consoleLogging   Pretty print logging
      --debug            Set debug level
  -h, --help             help for grpcserver

Use "grpcserver [command] --help" for more information about a command.
```

## Installing gRPC cli

```
$ brew tap grpc/grpc
$ brew install --with-plugins grpc
```