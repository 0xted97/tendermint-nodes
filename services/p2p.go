package services

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/me/dkg-node/config"
)

type P2PService struct {
	ctx  context.Context
	host host.Host
}

func NewP2PService(ctx context.Context) *P2PService {
	return &P2PService{ctx: ctx}
}

func (p *P2PService) Name() string {
	return "p2p"
}

func (p *P2PService) OnStart() error {
	// Random private key, will be modify after
	prvKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return err
	}
	fmt.Printf("prvKey: %v\n", prvKey)
	// prvKey, err := crypto.UnmarshalSecp256k1PrivateKey(make([]byte, 128))
	// if err != nil {
	// 	return err
	// }
	fmt.Printf("config.GlobalConfig.P2PAddress: %v\n", config.GlobalConfig.P2PAddress)

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(config.GlobalConfig.P2PAddress),
		libp2p.Identity(prvKey),
		libp2p.DisableRelay(),
	)

	if err != nil {
		return fmt.Errorf("failed to create p2p host: %w", err)
	}
	p.host = h

	// Print the host's PeerInfo in multiaddr format
	fmt.Printf("P2P host started with ID %s and address %s\n", h.ID(), h.Addrs()[0])

	return nil
}

func (p *P2PService) OnStop() error {
	// Close the libp2p host and its associated connections
	if p.host != nil {
		return p.host.Close()
	}
	return nil
}
