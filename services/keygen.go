package services

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	dkgSendProtocolID    = "/dkg/send/1.0.0"
	dkgReceiveProtocolID = "/dkg/receive/1.0.0"
)

type KeyGenService struct {
	ctx context.Context
	dkg *DKG

	p2p *P2PService
}

func NewKeyGenService(ctx context.Context, p2p *P2PService) *KeyGenService {
	return &KeyGenService{
		ctx: ctx,
		p2p: p2p,
		dkg: InitializeDKG(5, 3),
	}
}

func (p *KeyGenService) Name() string {
	return "keygen"
}

func (k *KeyGenService) Initialize() {
	k.p2p.host.SetStreamHandler(protocol.ID(dkgSendProtocolID), k.handleDKGSendStream)
	k.p2p.host.SetStreamHandler(protocol.ID(dkgReceiveProtocolID), k.handleDKGReceiveStream)
}

func (k *KeyGenService) OnStart() error {
	// Initial event
	k.Initialize()
	k.GenerateAndSendShares()
	// Test, it will remove after
	// peerID, _ := peer.Decode(k.p2p.peers[len(k.p2p.peers)-1].ID.Pretty())
	// k.p2p.SendMessage(peerID, dkgSendProtocolID, []byte("Send DKG"))
	// k.p2p.SendMessage(peerID, dkgReceiveProtocolID, []byte("Receive DKG"))
	return nil
}

func (k *KeyGenService) handleDKGSendStream(stream network.Stream) {
	// Implement the logic for handling the DKG send stream
	// ...
}

func (k *KeyGenService) handleDKGReceiveStream(stream network.Stream) {
	defer stream.Close()

	buf := make([]byte, 1024) // Adjust buffer size according to your needs
	n, err := stream.Read(buf)
	if err != nil {
		fmt.Printf("Error reading from DKG receive stream: %v\n", err)
		return
	}

	receivedMessage := buf[:n]
	fmt.Printf("Received DKG message: %s\n", string(receivedMessage))

	// Process the received DKG message
	// ...
}

func (k *KeyGenService) GenerateAndSendShares() error {
	privateKeyShare, _, _ := GenerateShares(k.dkg.n, k.dkg.t)
	fmt.Printf("privateKeyShare: %v\n", privateKeyShare)
	for i, peer := range k.p2p.peers {
		fmt.Printf("peer: %v\n", peer)
		shareMessage, _ := EncodeShare(i, privateKeyShare)
		if err := k.p2p.SendMessage(peer.ID, dkgSendProtocolID, shareMessage); err != nil {
			return fmt.Errorf("failed to send DKG share to node %d: %w", i, err)
		}
	}

	return nil
}

func (p *KeyGenService) OnStop() error {
	return nil
}
