# Fractal Engine (Tokenisation Engine)
This engine runs alongside the DOGE Layer 1 and allows users to mint tokens, and buy and sell tokens between each other.

## Automated Testing

### Fractal Engine e2e test with 1 Node (no gossip)
This test does the following: 
- Spins up a Doge Core instance
- Sets up a few wallets with account balances
- Calls the 'create mint' RPC API on Fractal Engine
- Sends the new mint hash to the Doge Core
- Waits for a 'Mint' to be validated and confirmed

### Running Unit Tests
`go test -v .\internal\test\unit\...`

### Running E2E Tests
Ensure Docker is running (WSL2 active for windows).

`go test -v -parallel=1 .\internal\test\e2e\...`

## Generate Protobuffers

`protoc --proto_path=. --go_out=. .\pkg\protocol\mint.proto`
`protoc --proto_path=. --go_out=. .\pkg\protocol\sell_offers.proto`
`protoc --proto_path=. --go_out=. .\pkg\protocol\buy_offers.proto`
`protoc --proto_path=. --go_out=. .\pkg\protocol\invoices.proto`
`protoc --proto_path=. --go_out=. .\pkg\protocol\payment.proto`

## Flows

- DogeFollower: L1 messages are listened too and if FE messages are identified they are stored into onchain_transactions. As soon as there are FE messages identified, it will not seek the next block until they have been matched by the MatcherService.
- DogeNetClient: DogeNet gossips (for mints) are stored into unconfirmed_mints.
- Processor: This will continuely run and attempt to match unconfirmed_mints with onchain_transactions, if matched correctly the onchain_transaction is removed and the mint is moved from unconfirmed_mints to mints.
- TrimmerService: Periodically removes old unconfirmed_mints.

## Docs

```sh
scripts/generate_docs.sh
```

## Generate Swagger Docs
`swag init --parseDependency --parseInternal --parseDepth 1 -g pkg/rpc/server.go`

## Running Fractal Engine
### Example Docker Compose
There is an example of running Dogecoin Core, DogeNet, Fractal Engine, and Fractal Admin UI.
You can just run `docker compose up`.

### Docker Compose with deps only (dogenet, dogecoin core)
This is usually if you want to dev on fractal engine or UI locally and its simple to just spin up a DogeNet + Dogecoin Core (regtest).
`docker compose --profile deps up`
Note: May require dockerfile configurations.

If you are running DogeNet in container and running Fractal Engine locally you can run it like this.
`go run cmd/fractal-engine/fractal_engine.go --doge-net-network tcp --doge-net-address localhost:8085`
By default it expects to use a unix socket, but if its in a container we need to connect via TCP.

### Docker Compose with fractal only (fractal engine, fractal ui)
This is usually if you want to run fractal against your existing DogeNet and Dogecoin Core instances.
`docker compose --profile fractal up`
Note: May require dockerfile configurations.