# Fractal Engine
This is a 'side chain' to give DOGE the support for fractionalised tokens that can be used for RWA or for NFTs.

## Architecture
- HTTP API for serving a Restful API for interacting with the Fractal Engine
- RPC Client for 'listening' for changes on the Doge L1 and for performing RPC calls to verify tokens.
- DB store is currently sqlite for local and postgres for production

## Doge Foundation Dependancies
- This project uses Chainfollower this is a library used for listening to blocks/syncing blocks from the Doge L1.

## Packages
### API
This is the HTTP API for consumers to interact with the Fractal Engine.

### Client
This is the Fractal Engine client consumers can use to interact with the FE API.

### Doge
Any Doge specific logic or operations live in this package.

### Protocol
This is how we encode/decode things from the FE onto/from the Doge L1.

### Store
This is how we store persistent data for the FE.

## Running
`go run cmd/fractal/fractal.go` will run the process that will listen for HTTP API requests and listen for changes and call the Doge RPC L1 when required.

## Manual/Interactive Testing
`go run cmd/tester/tester.go` has snippets in for running manual interactive tests with the fractal engine.

## Automated Tests
`go test -v .` when the tests are written will run those :)
