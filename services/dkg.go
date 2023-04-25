package services

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"encoding/hex"
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

func CombineSecrets(secrets [][]byte) ([]byte, error) {
	// Concatenate the secrets together
	combinedSecret := []byte{}
	for _, secret := range secrets {
		combinedSecret = append(combinedSecret, secret...)
	}

	// Compute the corresponding public key from the secret key
	privateKey, err := crypto.ToECDSA(combinedSecret)
	if err != nil {
		return nil, err
	}
	publicKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)

	return publicKeyBytes, nil
}

func CombinePublicKey(publicKeys [][]byte) ([]byte, error) {
	if len(publicKeys) < 2 {
		return nil, errors.New("need at least two public keys to combine")
	}

	// Parse the public keys
	publicKeysECDSA := make([]*ecdsa.PublicKey, len(publicKeys))
	for i, publicKey := range publicKeys {
		publicKeyECDSA, err := crypto.UnmarshalPubkey(publicKey)
		if err != nil {
			return nil, err
		}
		publicKeysECDSA[i] = publicKeyECDSA
	}

	// Add the x values of the public keys together
	x := new(big.Int)
	for _, publicKey := range publicKeysECDSA {
		x = x.Add(x, publicKey.X)
	}

	// Compute the corresponding y value using the elliptic curve equation
	curve := crypto.S256()
	ySquared := new(big.Int).Exp(x, big.NewInt(3), curve.Params().P)
	ySquared.Add(ySquared, curve.Params().B)
	y := new(big.Int).SetBit(ySquared, ySquared.BitLen(), 1) // set the highest bit
	y = new(big.Int).ModSqrt(y, curve.Params().P)
	if y == nil {
		return nil, fmt.Errorf("invalid public keys")
	}

	// Create a new public key using the combined x and y values
	publicKeyECDSA := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	// Encode the resulting public key as a byte slice
	masterPublicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	return masterPublicKeyBytes, nil
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
func TestPublicKey() error {
	priv1, public1, _, _ := GenerateShares(3, 2)
	priv2, public2, _, _ := GenerateShares(3, 2)
	priv3, public3, _, _ := GenerateShares(3, 2)
	var pubs = [][]byte{public1, public2, public3}
	public, err := CombinePublicKey(pubs)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		return err
	}
	fmt.Printf("public: %v\n", public)
	fmt.Printf("hex.EncodeToString(public): %v\n", hex.EncodeToString(public))

	var prvis = [][]byte{priv1, priv2, priv3}
	priv, err := CombineSecrets(prvis)
	fmt.Printf("hex.EncodeToString(priv): %v\n", hex.EncodeToString(priv))
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
