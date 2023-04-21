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

	p2p     *P2PService
	abciApp *ABCIApp
}

func NewKeyGenService(ctx context.Context) *KeyGenService {
	return &KeyGenService{
		ctx: ctx,
		// p2p: p2p,
		dkg: InitializeDKG(5, 3),
	}
}

func (p *KeyGenService) Name() string {
	return "keygen"
}

func (k *KeyGenService) InjectServices(p2p *P2PService, abciApp *ABCIApp) {
	k.p2p = p2p
	k.abciApp = abciApp
	k.Initialize()
}

func (k *KeyGenService) Initialize() {
	k.p2p.host.SetStreamHandler(protocol.ID(dkgSendProtocolID), k.handleDKGSendStream)

	// k.GenerateAndSendShares()
}

func (k *KeyGenService) OnStart() error {
	// Initial event

	// Test, it will remove after
	// peerID, _ := peer.Decode(k.p2p.peers[len(k.p2p.peers)-1].ID.Pretty())
	// k.p2p.SendMessage(peerID, dkgSendProtocolID, []byte("Send DKG"))
	// k.p2p.SendMessage(peerID, dkgReceiveProtocolID, []byte("Receive DKG"))
	return nil
}

func (k *KeyGenService) handleDKGSendStream(stream network.Stream) {
	defer stream.Close()
	buf := make([]byte, 128)
	n, err := stream.Read(buf)
	if err != nil {
		fmt.Printf("Error reading from stream: %v\n", err)
		return
	}
	share, err := DecodeShare(buf[:n])
	if err != nil {
		fmt.Println("Error decoding share:", err)
		return
	}
	k.abciApp.state.ReceiveShares[share.ID] = []Share{share}
	fmt.Printf("share: %v\n", share.ID)
}

func (k *KeyGenService) GenerateAndSendShares() error {
	for si := 0; si < 100; si++ {
		secret, shares, _ := GenerateShares(k.dkg.n, k.dkg.t)
		for i, peer := range k.p2p.peers {
			node := k.p2p.peersDetail[i]
			shareMessage, _ := EncodeShare(si, node.Index, shares[node.Index])
			if node.Self {
				continue
			}
			if err := k.p2p.SendMessage(peer.ID, dkgSendProtocolID, shareMessage); err != nil {
				return fmt.Errorf("failed to send DKG share to node %d: %w", i, err)
			}
		}
		k.abciApp.state.SecretShare = append(k.abciApp.state.SecretShare, secret)
	}
	return nil
}

func (p *KeyGenService) OnStop() error {
	return nil
}
