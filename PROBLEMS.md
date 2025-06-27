# Problems

- Mint to L1 without any metadata
Add a timeout for the mint

- Mint metdata without any L1 
Add a timeout for the mint

- Sell offer spam
Add a limit of sell offers per seller

- Buy offer spam
Add a limit of buy offers per buyer

- Invoice spam
Add a limit of invoices per seller

- Invoice to L1 without any metadata
Add a timeout

- Invoice to metadata without any L1
Add a timeout

- Fake invoice from attacker
The seller needs to sign the invoice

- Node gets L1 block 4 (missing metdata), then L1 block 5 (has metadata for)
For mints this is fine.
For sell offers/buy offers if refering to an unknown mint we need to ask and set a timeout of N
For invoices if refering to an unknown mint/offer we need to ask and set a timeout of N
For payment if refering to an unknown invoice we need to ask and set a timeout of N

