# Quick Start

## Full Setup
This is an example of running Dogecoin Core, DogeNet, Fractal Engine, and Fractal Admin UI.
```
docker compose up
```

## Deps only
This is an example of running Dogecoin Core, DogeNet in a container and you running Fractal Engine and/or Fractal Admin UI locally for development purposes.
```
docker compose --profile deps up
```

```
go run cmd/fractal-engine/fractal_engine.go --doge-net-network tcp --doge-net-address localhost:8085
```
Note: We override the DogeNet settings because in the full docker compose setup it uses a socket file, but if you are running DogeNet in a container and Fractal Engine locally you will need to connect too via TCP or you will need to mount a volume that lets you reference the socket file.

## Fractal only
This is an example of running Fractal Engine and Fractal Admin UI inside of a docker container. You will also have DogeNet and Dogecoin Core running somewhere else already.
```
docker compose --profile deps up
```

```
go run cmd/fractal-engine/fractal_engine.go --doge-net-network CHANGEME --doge-net-address CHANGEME --doge-scheme CHANGEME --doge-host CHANGEME --doge-port CHANGEME --doge-user CHANGEME --doge-password CHANGME
```