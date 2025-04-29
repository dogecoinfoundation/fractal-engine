# Fractal Engine flows

## Minting

- User calls 'mint' on FE API
- FE API writes 'unconfirmed mint' to DB and returns a 'hash'
- User writes transaction (to self - fees) referencing with `mint:<hash>` in an OP_RETURN
- FE API listens (using chain follower) to the L1
- When a mint is received (`mint:<hash>`) we store it in the db to be checked against the 'unconfirmed mints'
- This is done on another process that keeps checking (this handles the case if it gets a mint on L1 but hasnt synced the unconfirmed mint from the peer nodes)
- Once a mint is confirmed we can update the 'account' table that is a rollup of who owns which assets

## Offers
### Making offers
- User calls 'make offer' on FE API
- FE API writes to a 'unconfirmed offers' table in the DB and returns a 'hash'
- User writes transaction (to self - fees) referencing with `offer:<hash>` in an OP_RETURN
- When an offer is received (`offer:<hash>`) we store it in the db to be checked against the 'unconfirmed offers'
- This is done on another process that keeps checking (this handles the case if it gets an offer on L1 but hasnt synced the data from the peer nodes)

### Revoking offers
- User calls 'revoke offer' on FE API
- FE API writes to a 'unconfirmed revoked offers' table in the DB and returns a 'hash'
- User writes transaction (to self - fees) referencing with `revokeoffer:<hash>` in an OP_RETURN
- When an offer is received (`revokeoffer:<hash>`) we store it in the db to be checked against the 'unconfirmed revoked offers'
- It removes revoked offer from 'offers' table.
- This is done on another process that keeps checking (this handles the case if it gets an offer on L1 but hasnt synced the data from the peer nodes)

## Buying/Selling

- User hits '/offers' endpoint to review offers (with search params + pagination)
- An offer contains addresses of nodes (limited to a few) that are 'owners' of the token
- Once an offer is found user hits FE API (using an owner address) with 'accept offer'
- This is to ensure that only 1 person can accept the offer
- Offer is updated in DB with who accepted the offer
- FE API listens to L1 for a payment that comes in with the `offer:<hash>` so it can confirm the offer is paid for.
- FE API then updates the 'account' table with the newly reflected token ownership
- We could actually have a 'ledger' that we calculate the 'account' table from, so we have an audit trail of token movements.

- We could also have a process that runs and checks an offer and if it reaches the 'expiry' before being paid for, it reverts back to being an unaccepted offer.


