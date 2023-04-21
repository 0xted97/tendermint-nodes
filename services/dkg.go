package services

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"encoding/json"

	"github.com/hashicorp/vault/shamir"
)

type KeyGenMessage struct {
	ID        int
	PublicKey []byte
	Share     Share
}

type Share struct {
	NodeIndex int
	Share     []byte
}

func EncodeShare(id int, nodeIndex int, share []byte) ([]byte, error) {
	shareData := Share{
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
func GenerateShares(n int, t int) ([]byte, []byte, [][]byte, error) {
	curve := elliptic.P256()
	// Generate a random secret
	secret, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate random secret: %w", err)
	}

	// Compute the public key from the private key
	publicKey := elliptic.Marshal(curve, x, y)

	// Create shares using Shamir's Secret Sharing
	shares, err := shamir.Split(secret, n, t)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate shares: %w", err)
	}

	return secret, publicKey, shares, nil
}
