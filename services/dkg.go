package services

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
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

func CombineSecrets(secrets [][]byte) (ecdsa.PrivateKey, ecdsa.PublicKey, error) {
	// Convert each secret to an ECDSA private key
	privateKeys := make([]*ecdsa.PrivateKey, len(secrets))
	// publicKeys := make([]*ecdsa.PublicKey, len(secrets))
	curve := crypto.S256() // secp256k1
	for i, secret := range secrets {
		privateKey, err := crypto.ToECDSA(secret)
		if err != nil {
			return ecdsa.PrivateKey{}, ecdsa.PublicKey{}, err
		}
		privateKeys[i] = privateKey
	}

	// Add the private keys together
	totalPrivateKey := ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: crypto.S256(),
		},
		D: &big.Int{},
	}
	for _, privateKey := range privateKeys {
		totalPrivateKey.D = new(big.Int).Mod(new(big.Int).Add(totalPrivateKey.D, privateKey.D),
			curve.Params().N)
	}

	// totalPrivateKey.D = totalPrivateKey.D.Mod(totalPrivateKey.D, crypto.S256().Params().N)
	totalPrivateKey.PublicKey.X, totalPrivateKey.PublicKey.Y = totalPrivateKey.PublicKey.Curve.ScalarBaseMult(totalPrivateKey.D.Bytes())
	totalPublicKey := totalPrivateKey.PublicKey
	return totalPrivateKey, totalPublicKey, nil
}

func CombinePublicKey(publicKeys [][]byte) (ecdsa.PublicKey, error) {
	if len(publicKeys) < 2 {
		return ecdsa.PublicKey{}, errors.New("need at least two public keys to combine")
	}
	curve := crypto.S256() // secp256k1
	// Convert each public key to a *secp256k1.PublicKey
	ecdsaPublicKeys := make([]*ecdsa.PublicKey, len(publicKeys))
	for i, publicKey := range publicKeys {
		var err error
		ecdsaPublicKeys[i], err = crypto.UnmarshalPubkey(publicKey)
		if err != nil {
			return ecdsa.PublicKey{}, err
		}
	}
	totalPublicKey := ecdsa.PublicKey{
		Curve: curve,
		X:     &big.Int{},
		Y:     &big.Int{},
	}

	for _, public := range ecdsaPublicKeys {
		totalPublicKey.X, totalPublicKey.Y = totalPublicKey.Add(totalPublicKey.X, totalPublicKey.Y, public.X, public.Y)
	}
	return totalPublicKey, nil
}

type DKG struct {
	n int
	t int
}

func TestPublicKey() error {
	priv1, public1, _, _ := GenerateShares(3, 2)
	priv2, public2, _, _ := GenerateShares(3, 2)
	priv3, public3, _, _ := GenerateShares(3, 2)

	var prvis = [][]byte{priv1, priv2, priv3}
	_, p1, err := CombineSecrets(prvis)
	if err != nil {
		return err
	}

	var pubs = [][]byte{public1, public2, public3}

	p2, err := CombinePublicKey(pubs)
	if err != nil {
		return err
	}

	if p1.X.Cmp(p2.X) == 0 && p1.Y.Cmp(p2.Y) == 0 {
		fmt.Println("Both methods produce the same result.")
	} else {
		fmt.Println("Results are different.")
	}

	return nil
}

// GenerateShares generates a random secret and creates n shares with a threshold of t.
// Returns the generated secret, and a slice of shares.
func GenerateShares(n int, t int) ([]byte, []byte, [][]byte, error) {
	// Generate a random secret key on the secp256k1 curve
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, nil, nil, err
	}
	// Compute the corresponding public key
	publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)
	// Split the secret key into shares
	shares, err := shamir.Split(privateKey.D.Bytes(), n, t)
	if err != nil {
		return nil, nil, nil, err
	}

	return privateKey.D.Bytes(), publicKeyBytes, shares, nil
}
