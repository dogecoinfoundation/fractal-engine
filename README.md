# Fractal Engine (Tokenisation Engine)
This engine runs alongside the DOGE Layer 1 and allows users to mint tokens, and buy and sell tokens between each other.

## Architecture
[See architecture docs here](ARCHITECTURE.md)


## Generate Protobuffers

`protoc --proto_path=. --go_out=. .\pkg\protocol\mint.proto`

## Services

- DogeFollower: L1 messages are listened too and if FE messages are identified they are stored into onchain_transactions. As soon as there are FE messages identified, it will not seek the next block until they have been matched by the MatcherService.
- DogeNetClient: DogeNet gossips (for mints) are stored into unconfirmed_mints.
- MatcherService: This will continuely run and attempt to match unconfirmed_mints with onchain_transactions, if matched correctly the onchain_transaction is removed and the mint is moved from unconfirmed_mints to mints.
- TrimmerService: Periodically removes old unconfirmed_mints.

## Flows

- Count onchain_transactions for block_height == 0 -> Read Block from L1 -> Store in onchain_transactions
- REST API POST /mints -> store into unconfirmed_mints
- Receive Mint Gossip -> store into unconfirmed_mints
- Every 5 seconds -> Match unconfirmed_mints with onchain_transactions

 Note: Maybe instead of sending just confirmed, send unconfirmed because someone could attack the Origin Node.