package store

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
)

func (s *TokenisationStore) GetSellOffers(offset int, limit int, mintHash string, offererAddress string) ([]SellOffer, error) {
	var rows *sql.Rows
	var err error

	log.Println("GetSellOffers", mintHash, offererAddress, limit, offset)

	if offererAddress != "" {
		rows, err = s.DB.Query("SELECT id, created_at, offerer_address, hash, mint_hash, quantity, price, public_key FROM sell_offers WHERE mint_hash = $1 AND offerer_address = $2 LIMIT $3 OFFSET $4", mintHash, offererAddress, limit, offset)
	} else {
		rows, err = s.DB.Query("SELECT id, created_at, offerer_address, hash, mint_hash, quantity, price, public_key FROM sell_offers WHERE mint_hash = $1 LIMIT $2 OFFSET $3", mintHash, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []SellOffer

	for rows.Next() {
		var offer SellOffer
		if err := rows.Scan(&offer.Id, &offer.CreatedAt, &offer.OffererAddress, &offer.Hash, &offer.MintHash, &offer.Quantity, &offer.Price, &offer.PublicKey); err != nil {
			return nil, err
		}

		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}

func (s *TokenisationStore) CountSellOffers(mintHash string, offererAddress string) (int, error) {
	row := s.DB.QueryRow("SELECT COUNT(*) FROM sell_offers WHERE mint_hash = $1 AND offerer_address = $2", mintHash, offererAddress)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *TokenisationStore) SaveSellOffer(d *SellOfferWithoutID) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO sell_offers (id, offerer_address, hash, mint_hash, quantity, price, created_at, public_key, signature)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, d.OffererAddress, d.Hash, d.MintHash, d.Quantity, d.Price, d.CreatedAt, d.PublicKey, d.Signature)

	return id, err
}

func (s *TokenisationStore) DeleteSellOffer(hash string, publicKey string) error {
	_, err := s.DB.Exec("DELETE FROM sell_offers WHERE hash = $1 AND public_key = $2", hash, publicKey)
	return err
}
