package services

import (
	"context"

	"github.com/me/dkg-node/auth"
)

type VerifierService struct {
	verifier auth.GeneralVerifier
	ctx      context.Context
}

func NewVerifierService(ctx context.Context) *VerifierService {
	return &VerifierService{
		ctx: ctx,
	}
}

func (v *VerifierService) Name() string {
	return "auth"
}

func (v *VerifierService) OnStart() error {
	verifiers := []auth.Verifier{
		auth.NewGoogleVerifier(),
		auth.NewFacebookVerifier(),
	}

	v.verifier = auth.NewAuthService(verifiers)
	return nil
}

func (k *VerifierService) OnStop() error {
	return nil
}
