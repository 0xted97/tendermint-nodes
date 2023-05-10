package services

import (
	"context"

	"github.com/me/dkg-node/auth"
)

type VerifierService struct {
	verifier auth.GeneralVerifier
	ctx      context.Context
}

func NewVerifierService(services *Services) (*VerifierService, error) {
	verifierService := &VerifierService{ctx: services.Ctx}
	verifiers := []auth.Verifier{
		auth.NewGoogleVerifier(),
		auth.NewFacebookVerifier(),
	}

	verifierService.verifier = auth.NewAuthService(verifiers)
	services.VerifierService = verifierService
	return verifierService, nil
}

func (v *VerifierService) Name() string {
	return "auth"
}

func (k *VerifierService) OnStop() error {
	return nil
}
