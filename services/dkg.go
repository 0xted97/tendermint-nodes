package services

import (
	"crypto/rand"
	"fmt"

	"encoding/json"

	"github.com/hashicorp/vault/shamir"
)

type Share struct {
	ID    int
	Share []byte
}

func EncodeShare(id int, share []byte) ([]byte, error) {
	shareData := Share{
		ID:    id,
		Share: share,
	}

	encodedShare, err := json.Marshal(shareData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode share: %w", err)
	}

	return encodedShare, nil
}

type DKG struct {
	n int
	t int
}

func InitializeDKG(n int, t int) *DKG {
	return &DKG{
		n: n,
		t: t,
	}
}

// GenerateShares generates a random secret and creates n shares with a threshold of t.
// Returns the generated secret, and a slice of shares.
func GenerateShares(n int, t int) ([]byte, [][]byte, error) {
	// Generate a random secret
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Create shares using Shamir's Secret Sharing
	shares, err := shamir.Split(secret, n, t)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate shares: %w", err)
	}

	return secret, shares, nil
}
