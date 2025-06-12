# Fractal Engine (Tokenisation Engine)
This engine runs alongside the DOGE Layer 1 and allows users to mint tokens, and buy and sell tokens between each other.

## Architecture
[See architecture docs here](ARCHITECTURE.md)


## Generate Protobuffers

`protoc --proto_path=. --go_out=. .\pkg\protocol\mint.proto`

## Flows

- DogeFollower: L1 messages are listened too and if FE messages are identified they are stored into onchain_transactions. As soon as there are FE messages identified, it will not seek the next block until they have been matched by the MatcherService.
- DogeNetClient: DogeNet gossips (for mints) are stored into unconfirmed_mints.
- MatcherService: This will continuely run and attempt to match unconfirmed_mints with onchain_transactions, if matched correctly the onchain_transaction is removed and the mint is moved from unconfirmed_mints to mints.
- TrimmerService: Periodically removes old unconfirmed_mints.