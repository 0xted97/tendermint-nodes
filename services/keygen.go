package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

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
		dkg: InitializeDKG(3, 2),
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
	var buf bytes.Buffer
	_, err := io.Copy(&buf, stream)
	if err != nil {
		fmt.Printf("Error reading from stream: %v\n", err)
		return
	}
	var keyGenMessage KeyGenMessage
	err = json.Unmarshal(buf.Bytes(), &keyGenMessage)
	if err != nil {
		fmt.Println("failed to unmarshal keygen message: %w", err)
		return
	}
	shares := k.abciApp.state.ReceiveShares[keyGenMessage.ID]
	k.abciApp.state.ReceiveShares[keyGenMessage.ID] = append(shares, keyGenMessage.Share)
	k.abciApp.state.ReceivePublicKeys[keyGenMessage.ID] = append(k.abciApp.state.ReceivePublicKeys[keyGenMessage.ID], keyGenMessage.PublicKey)

	k.abciApp.SaveState()
}

func (k *KeyGenService) GenerateAndSendShares() error {
	keysLength := 5
	for si := 0; si < keysLength; si++ {
		secret, publicKey, shares, _ := GenerateShares(k.dkg.n, k.dkg.t)
		for i, peer := range k.p2p.peers {
			node := k.p2p.peersDetail[i]
			share := Share{
				NodeIndex: node.Index,
				Share:     shares[node.Index],
			}
			keyGenMessage := KeyGenMessage{
				ID:        si,
				PublicKey: publicKey,
				Share:     share,
			}

			keyGenMessageByte, err := json.Marshal(keyGenMessage)
			if err != nil {
				return fmt.Errorf("failed to encode keygen message: %w", err)
			}
			if node.Self {
				// Set my secret
				shares := k.abciApp.state.ReceiveShares[keyGenMessage.ID]
				k.abciApp.state.ReceiveShares[keyGenMessage.ID] = append(shares, share)
				k.abciApp.state.ReceivePublicKeys[keyGenMessage.ID] = append(k.abciApp.state.ReceivePublicKeys[keyGenMessage.ID], keyGenMessage.PublicKey)
			} else {
				if err := k.p2p.SendMessage(peer.ID, dkgSendProtocolID, keyGenMessageByte); err != nil {
					return fmt.Errorf("failed to send DKG share to node %d: %w", i, err)
				}
			}
			// Store share that sent to the other nodes
			k.abciApp.state.SentShares[keyGenMessage.ID] = append(k.abciApp.state.SentShares[keyGenMessage.ID], share)
		}
		k.abciApp.state.SecretShare[si] = secret
	}
	// Initial state first
	k.abciApp.state.LastCreatedIndex = keysLength
	k.abciApp.state.LastUnassignedIndex = 0
	// Save state
	k.abciApp.SaveState()
	return nil
}

func (k *KeyGenService) OnStop() error {
	return nil
}
