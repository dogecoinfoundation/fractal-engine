

Prepare Mint
- Post request to prepare mint
- It protobuf packs the request and returns base64 encoded message (encoded_request)
- It encodes the transaction body to be written to OP_RETURN on L1 (encoded_transaction_body)

Sign Tx (this can be used by consumers that dont have to sign the transaction themselves)
- Post request to sign tx (passing in private key)
- It calls L1 endpoint signtransactionwithkey (using private key)
- It returns the signed transaction (signed_transaction)

Create Mint
- Post Payload (with Signature validation) + Signed Transaction



fundrawtransaction
signrawtransactionwithwallet
