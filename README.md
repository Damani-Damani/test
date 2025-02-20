# Control Server

Controlserver is a server that facilitates Clearbot communication with the frontend.

## Architecture

- The robot <-> server conection is made over gRPC with streaming connections.
- The server <-> user connection is made over websocet.

## Development

### Making changes to the gRPC definitions

The proto files are present in the [proto](/proto) folder. See [](/proto/controlserver.proto)

Once done, run the generator:

```
protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. proto/controlserver.proto
```

### Running the application

The project has `air` configured.

```
air -c .air.toml
```



