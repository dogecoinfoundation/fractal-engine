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
`protoc --proto_path=. --go_out=. .\pkg\protocol\offers.proto`
`protoc --proto_path=. --go_out=. .\pkg\protocol\invoices.proto`
`protoc --proto_path=. --go_out=. .\pkg\protocol\payment.proto`

## Flows

- DogeFollower: L1 messages are listened too and if FE messages are identified they are stored into onchain_transactions. As soon as there are FE messages identified, it will not seek the next block until they have been matched by the MatcherService.
- DogeNetClient: DogeNet gossips (for mints) are stored into unconfirmed_mints.
- Processor: This will continuely run and attempt to match unconfirmed_mints with onchain_transactions, if matched correctly the onchain_transaction is removed and the mint is moved from unconfirmed_mints to mints.
- TrimmerService: Periodically removes old unconfirmed_mints.


## Known Issues

- Handle overflows of Mint API call (Current limit is 100, but we need to add logic incase an unmatched mint gets discarded because of the limit)
- Handle overflows of Onchain Transactions from L1 (Similar issue, there is currently no limit, but in theory over time this could fill up with junk)

## Docs

```sh
scripts/generate_docs.sh
```

## Generate Swagger Docs
`swag init --parseDependency --parseInternal --parseDepth 1 -g pkg/rpc/server.go`

### TODO 

- Sign invoices from seller
- Ensure we put in a transaction the confirming of a invoice
- Have invoice limit (unconfirmed)
