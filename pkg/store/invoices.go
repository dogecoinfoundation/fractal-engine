package store

import "github.com/google/uuid"

func (s *TokenisationStore) GetInvoices(offset int, limit int, mintHash string, offererAddress string) ([]Invoice, error) {
	rows, err := s.DB.Query("SELECT id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, created_at FROM invoices WHERE buy_offer_mint_hash = $1 AND buy_offer_offerer_address = $2 LIMIT $3 OFFSET $4", mintHash, offererAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(&invoice.Id, &invoice.Hash, &invoice.PaymentAddress, &invoice.BuyOfferOffererAddress, &invoice.BuyOfferHash, &invoice.BuyOfferMintHash, &invoice.BuyOfferQuantity, &invoice.BuyOfferPrice, &invoice.CreatedAt); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

func (s *TokenisationStore) SaveUnconfirmedInvoice(invoice *UnconfirmedInvoice) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO unconfirmed_invoices (id, hash, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, invoice.Hash, invoice.BuyOfferOffererAddress, invoice.BuyOfferHash, invoice.BuyOfferMintHash, invoice.BuyOfferQuantity, invoice.BuyOfferPrice, invoice.CreatedAt)

	return id, err
}

func (s *TokenisationStore) SaveInvoice(invoice *Invoice) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO invoices (id, hash, payment_address, buy_offer_offerer_address, buy_offer_hash, buy_offer_mint_hash, buy_offer_quantity, buy_offer_price, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, invoice.Hash, invoice.PaymentAddress, invoice.BuyOfferOffererAddress, invoice.BuyOfferHash, invoice.BuyOfferMintHash, invoice.BuyOfferQuantity, invoice.BuyOfferPrice, invoice.CreatedAt)

	return id, err
}
