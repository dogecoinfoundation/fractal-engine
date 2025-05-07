# Goals

API -> Return Trxn


- Easy adoption
- Scalable
- Accessible
- Open

# Features

- Seller does not require a node to be able to sell. aka they can just post offers onto any node.
- Seller does not need to be online (to sign a trxn) to approve a sale.
- Buyer can connect to any node, query offers, and attempt to accept an offer.
- Double sell prevents from users wasting money, (ie. escrow refund, or some kind of locking mechanism)

# High Level Flow

- User mints token
-- Mint action gossiped to other FE nodes
-- Mint action is recorded onto the L1 (for anchoring/validation)
- Owner of token provides offers
-- Offer action is gossiped to other FE nodes
-- Offer action is recorded onto the L1
- Buyer can query offers
-- API call done on the FE
- Buyer accepts offer
-- Accept offer action gossiped to other FE nodes
-- Accept offer action is recorded onto the L1
-- Once an offer is verified on the L1 (figure out how to lock this? come to a concensus????) this is the big unknown atm ??????

-- Potentials
--- Lock/Mutex like a Raft concensus system
--- Or watch for L1 for payment (first in best dressed.)

-- Reservation system with concensus.

- Accept Offer Request/Reservation
- We use concensus to determine who got there first
- Then we know who to look for a payment for, plus we can let the end user know where to go ahead with the payment
- This removes a race and a refund flow that may be needed.

- How do we want to handle L1 rollbacks? We can wait for a number of blocks before marking it as 'final'?
- and if it rollbacks the payment, then we can update the status of the token.
