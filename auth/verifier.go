package auth

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Verifier interface {
	GetIdentifier() string
	CleanToken(string) string
	VerifyRequestIdentity(*json.RawMessage) (verified bool, verifierID string, err error)
}

type GeneralVerifier interface {
	ListVerifiers() []string
	Verify(*json.RawMessage) (verified bool, verifierID string, err error)
	Lookup(string) (Verifier, error)
}

type AuthService struct {
	Verifiers map[string]Verifier // a map of verifiers, keyed by social media platform name
}

type VerifyMessage struct {
	Token              string `json:"idtoken"`
	VerifierIdentifier string `json:"verifieridentifier"`
}

// ListVerifiers gets List of Registered Verifiers
func (tgv *AuthService) ListVerifiers() []string {
	list := make([]string, len(tgv.Verifiers))
	count := 0
	for k := range tgv.Verifiers {
		list[count] = k
		count++
	}
	return list
}

// Lookup returns the appropriate verifier
func (tgv *AuthService) Lookup(verifierIdentifier string) (Verifier, error) {
	if tgv.Verifiers == nil {
		return nil, errors.New("Verifiers mapping not initialized")
	}
	if tgv.Verifiers[verifierIdentifier] == nil {
		return nil, errors.New("Verifier with verifierIdentifier " + verifierIdentifier + " could not be found")
	}
	return tgv.Verifiers[verifierIdentifier], nil
}

func (tgv *AuthService) Verify(rawMessage *json.RawMessage) (bool, string, error) {
	var verifyMessage VerifyMessage
	if err := json.Unmarshal(*rawMessage, &verifyMessage); err != nil {
		return false, "", err
	}
	v, err := tgv.Lookup(verifyMessage.VerifierIdentifier)
	if err != nil {
		return false, "", err
	}
	cleanedToken := v.CleanToken(verifyMessage.Token)
	if cleanedToken != verifyMessage.Token {
		return false, "", errors.New("Cleaned token is different from original token")
	}
	return v.VerifyRequestIdentity(rawMessage)
}

// NewAuthService - Initialization function for a generic GeneralVerifier
func NewAuthService(verifiers []Verifier) GeneralVerifier {
	auth := &AuthService{
		Verifiers: make(map[string]Verifier),
	}
	for _, verifier := range verifiers {
		fmt.Printf("verifier.GetIdentifier(): %v\n", verifier.GetIdentifier())
		auth.Verifiers[verifier.GetIdentifier()] = verifier
	}
	return auth
}
