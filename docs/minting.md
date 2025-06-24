# Minting

```mermaid
sequenceDiagram
    Bob->>Fractal API: Mint my token
    Fractal->>Fractal Store: Write unconfirmed mint
    Fractal API->>DogeNet: Gossip mint
    Fractal API-->>Bob: Encoded mint payload to write on chain
    Bob->>L1: Write mint payload on chain
    L1->>Fractal Follower: On Mint FE Transaction
    Fractal Follower->>Fractal Store: Validate and confirm Mint
    Fractal Follower->>Fractal Store: Update Rollups
```