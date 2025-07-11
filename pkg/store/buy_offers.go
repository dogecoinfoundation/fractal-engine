package store

import (
	"database/sql"

	"github.com/google/uuid"
)

func (s *TokenisationStore) SaveBuyOffer(d *BuyOfferWithoutID) (string, error) {
	id := uuid.New().String()

	_, err := s.DB.Exec(`
	INSERT INTO buy_offers (id, offerer_address, seller_address, hash, mint_hash, quantity, price, created_at, public_key, signature)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, id, d.OffererAddress, d.SellerAddress, d.Hash, d.MintHash, d.Quantity, d.Price, d.CreatedAt, d.PublicKey, d.Signature)

	return id, err
}

func (s *TokenisationStore) CountBuyOffers(mintHash string, offererAddress string, sellerAddress string) (int, error) {
	row := s.DB.QueryRow("SELECT COUNT(*) FROM buy_offers WHERE mint_hash = $1 AND offerer_address = $2 AND seller_address = $3", mintHash, offererAddress, sellerAddress)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (s *TokenisationStore) DeleteBuyOffer(hash string, publicKey string) error {
	_, err := s.DB.Exec("DELETE FROM buy_offers WHERE hash = $1 AND public_key = $2", hash, publicKey)
	return err
}

func (s *TokenisationStore) GetBuyOffersByMintAndSellerAddress(offset int, limit int, mintHash string, sellerAddress string) ([]BuyOffer, error) {
	var rows *sql.Rows
	var err error

	if sellerAddress == "" {
		rows, err = s.DB.Query("SELECT id, created_at, offerer_address, seller_address, hash, mint_hash, quantity, price, public_key, signature FROM buy_offers WHERE mint_hash = $1 LIMIT $2 OFFSET $3", mintHash, limit, offset)
	} else {
		rows, err = s.DB.Query("SELECT id, created_at, offerer_address, seller_address, hash, mint_hash, quantity, price, public_key, signature FROM buy_offers WHERE mint_hash = $1 AND seller_address = $2 LIMIT $3 OFFSET $4", mintHash, sellerAddress, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []BuyOffer

	for rows.Next() {
		var offer BuyOffer
		if err := rows.Scan(&offer.Id, &offer.CreatedAt, &offer.OffererAddress, &offer.SellerAddress, &offer.Hash, &offer.MintHash, &offer.Quantity, &offer.Price, &offer.PublicKey, &offer.Signature); err != nil {
			return nil, err
		}

		offers = append(offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return offers, nil
}
