package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/me/dkg-node/config"
	"github.com/multiformats/go-multiaddr"
)

type P2PService struct {
	ctx  context.Context
	host host.Host

	peers       []peer.AddrInfo
	peersDetail []config.NodeDetail
}

func NewP2PService(ctx context.Context) *P2PService {
	return &P2PService{ctx: ctx}
}

func (p *P2PService) Name() string {
	return "p2p"
}

func (p *P2PService) OnStart() error {
	// Random private key, will be modify after
	// prvKey, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	// if err != nil {
	// 	return err
	// }

	ethPrivateKeyHex := config.GlobalConfig.NodePrivateKey
	ethPrivateKeyBytes, err := hex.DecodeString(ethPrivateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to decode Ethereum private key: %w", err)
	}

	prvKey, err := crypto.UnmarshalSecp256k1PrivateKey(ethPrivateKeyBytes)
	if err != nil {
		return err
	}

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

	if err := p.ConnectToPeers(); err != nil {
		return fmt.Errorf("failed to connect to peers: %w", err)
	}

	return nil
}

func (p *P2PService) ConnectToPeers() error {
	// Will get from smart contracts
	nodes := *config.NodeList
	for _, node := range nodes {

		peerAddr := node.P2PAddress
		addr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			return fmt.Errorf("invalid multiaddr: %w", err)
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return fmt.Errorf("failed to get AddrInfo: %w", err)
		}
		// check self address
		if strings.ToLower(node.EthAddress) == strings.ToLower(config.GlobalConfig.EthAddress) {
			p.peers = append(p.peers, *addrInfo)
			node.Self = true
			p.peersDetail = append(p.peersDetail, node)
			continue
		}

		err = p.host.Connect(p.ctx, *addrInfo)
		if err != nil {
			return fmt.Errorf("failed to connect to peer: %w", err)
		}
		// p.peer and p.connectedPeers are the same
		node.Self = false
		p.peers = append(p.peers, *addrInfo)
		p.peersDetail = append(p.peersDetail, node)

		fmt.Printf("Connected to peer %s\n", addrInfo.ID.Pretty())
	}

	return nil
}

func (p *P2PService) SendMessage(peerID peer.ID, protocolID protocol.ID, msg []byte) error {
	stream, err := p.host.NewStream(p.ctx, peerID, protocol.ID(protocolID))
	if err != nil {
		return fmt.Errorf("failed to create new stream: %w", err)
	}
	defer stream.Close()

	_, err = stream.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *P2PService) OnStop() error {
	// Close the libp2p host and its associated connections
	if p.host != nil {
		return p.host.Close()
	}
	return nil
}
