package services

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"encoding/json"

	"github.com/hashicorp/vault/shamir"
)

type Share struct {
	ID        int
	NodeIndex int
	Share     []byte
}

func EncodeShare(id int, nodeIndex int, share []byte) ([]byte, error) {
	shareData := Share{
		ID:        id,
		NodeIndex: nodeIndex,
		Share:     share,
	}

	encodedShare, err := json.Marshal(shareData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode share: %w", err)
	}

	return encodedShare, nil
}

func DecodeShare(encodedShare []byte) (Share, error) {
	var shareData Share

	err := json.Unmarshal(encodedShare, &shareData)
	if err != nil {
		return Share{}, fmt.Errorf("failed to encode share: %w", err)
	}

	return shareData, nil
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
	curve := elliptic.P256()
	// Generate a random secret
	secret, _, _, err := elliptic.GenerateKey(curve, rand.Reader)
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
