# Flows

## Common Flows

### Gossip Network (DogeNet)
We listen on the gossip network for metadata for mints, offers, invoices, payments.
We store each of these in their respective unconfirmed tables.
![Dogenet Follower](mermaid\0_common\dogenet_follower.png)

### Dogecoin L1 Follower
We listen for blocks on the Dogecoin L1. For each block we look at each transaction and look at the OP_RETURN to see if it is a fractal engine transaction. If it is we store this information into a onchain_transactions table.
![Onchain Transaction Follower](mermaid\0_common\onchain_transaction_follower.png)

## Minting

### Creating a Mint
A user calls the mint creation API which stores it in unconfirmed_mints and encodes an OP_RETURN transaction body with the message that Fractal Engine can understand, this is then returned to the user so they can sign it and write it to the Dogecoin L1.
![Mint Creation](mermaid\1_minting\1_mint_creation.png)

### Confirming a Mint
We periodically check the onchain transactions and unconfirmed mints, we then look for a match and validate these records.
Once we get a match we can move the unconfirmed mint into mints and update the token balance.
![Onchain Transaction Confirmation](mermaid\1_minting\2_onchain_transaction_confirmation.png)

## Offers and Invoices

### Sell Offers
A token owner can create a sell offer to let the potential buyers know what they want to sell these tokens for.
Offers are purely informational and nothing is actually binding.
![Sell Offers](mermaid\2_invoices_and_offers\1_sell_offers.png)

### Buy Offers
![Buy Offers](mermaid\2_invoices_and_offers\2_buy_offers.png)
A user who wants to buy tokens can view the sell offers and then make their owner offer.
This then allows the seller to pick what buy offer they like and create an invoice for it.

### Create Invoice
Once a token seller identifies an offer they like they can create an invoice. This is saved into unconfirmed invoices and written to Dogecoin L1.
![Create Invoice](mermaid\2_invoices_and_offers\3_create_invoice.png)

### Invoice Processing (no match)
When an invoice comes in on the Dogecoin L1, if there are enough tokens to be sold it will then create a pending token balance.
This flow example just shows this process happening (but does not show the match logic).
![Invoice Processing without match](mermaid\2_invoices_and_offers\4_invoice_processing_without_match.png)

### Invoice Processing (match)
When an unconfirmed invoice is matched to the counterpart on the Dogecoin L1 the invoice is validated and then moved into the invoice table. The payment address is then exposed so that the buyer can transfer a payment.
![Invoice Processing with match](mermaid\2_invoices_and_offers\5_invoice_processing_with_match.png)

## Payments

### Payment Processing
When a payment comes through on the Dogecoin L1 it then tries to match to an invoice. If matched, the inconfirmed payment is then moved into the payment table. The token balance for the seller is changed to reflect the transfer of tokens and the buyer token balance is updated. The pending token balance is then removed.
![Payment Processing](mermaid\3_payment\1_handle_payment.png)