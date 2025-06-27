package store

import "github.com/google/uuid"

func (s *TokenisationStore) GetOffers(offset int, limit int, mintHash string, typeInt int) ([]Offer, error) {
	rows, err := s.DB.Query("SELECT id, created_at, type, offerer_address, hash, mint_hash, quantity, price FROM offers WHERE mint_hash = $1 AND type = $2 LIMIT $3 OFFSET $4", mintHash, typeInt, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []Offer

	for rows.Next() {
		var offer Offer
		if err := rows.Scan(&offer.Id, &offer.CreatedAt, &offer.Type, &offer.OffererAddress, &offer.Hash, &offer.MintHash, &offer.Quantity, &offer.Price); err != nil {
			return nil, err
		}

		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (s *TokenisationStore) SaveOffer(d *OfferWithoutID) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO offers (id, type, offerer_address, hash, mint_hash, quantity, price, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, id, d.Type, d.OffererAddress, d.Hash, d.MintHash, d.Quantity, d.Price, d.CreatedAt)

	return id, err
}
