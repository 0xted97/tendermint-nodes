package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/me/dkg-node/config"
	"github.com/multiformats/go-multiaddr"
)

type P2PService struct {
	ctx             context.Context
	ethereumService *EthereumService
	host            host.Host

	peers       []peer.AddrInfo
	peersDetail []config.NodeDetail
}

func NewP2PService(services *Services) (*P2PService, error) {
	p2pService := &P2PService{ctx: services.Ctx, ethereumService: services.EthereumService}
	ethPrivateKeyHex := services.ConfigService.NodePrivateKey
	ethPrivateKeyBytes, err := hex.DecodeString(ethPrivateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Ethereum private key: %w", err)
	}

	prvKey, err := crypto.UnmarshalSecp256k1PrivateKey(ethPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.ListenAddrStrings(services.ConfigService.P2PAddress),
		libp2p.Identity(prvKey),
		libp2p.DisableRelay(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create p2p host: %w", err)
	}
	p2pService.host = h

	// Print the host's PeerInfo in multiaddr format
	fmt.Printf("P2P host started with ID %s and address %s\n", h.ID(), h.Addrs()[0])

	// if err := p.ConnectToPeers(); err != nil {
	// 	return fmt.Errorf("failed to connect to peers: %w", err)
	// }

	services.P2PService = p2pService
	return p2pService, nil
}

func (p *P2PService) Name() string {
	return "p2p"
}

func (p *P2PService) ConnectToPeer(nodeAddress common.Address) (NodeReference, error) {
	// TODO: get detail node by address from smart contracts
	node, err := p.ethereumService.NodeDetail(nodeAddress)
	if err != nil {
		return NodeReference{}, err
	}
	nodeRef := NodeReference{
		Address:   new(common.Address),
		Index:     &big.Int{},
		PeerID:    "",
		PublicKey: &ecdsa.PublicKey{},
		Self:      false,
	}
	peerAddr := node.P2PAddress
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return NodeReference{}, fmt.Errorf("invalid multiaddr: %w", err)
	}

	addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return NodeReference{}, fmt.Errorf("failed to get AddrInfo: %w", err)
	}
	*nodeRef.Address = common.HexToAddress(node.EthAddress)
	*nodeRef.Index = *big.NewInt(int64(node.Index))
	nodeRef.PeerID = addrInfo.ID
	pubBytes, _ := hex.DecodeString(node.EthPub)

	x, y := elliptic.Unmarshal(p.ethereumService.EthCurve, pubBytes)
	nodeRef.PublicKey = &ecdsa.PublicKey{
		Curve: p.ethereumService.EthCurve,
		X:     x,
		Y:     y,
	}
	// check self address
	if strings.ToLower(node.EthAddress) == strings.ToLower(config.GlobalConfig.EthAddress) {
		p.peers = append(p.peers, *addrInfo)
		p.peersDetail = append(p.peersDetail, node)

		nodeRef.Self = true
	} else {
		err = p.host.Connect(p.ctx, *addrInfo)
		if err != nil {
			return NodeReference{}, fmt.Errorf("failed to connect to peer: %w", err)
		}
		// p.peer and p.connectedPeers are the same
		p.peers = append(p.peers, *addrInfo)
		p.peersDetail = append(p.peersDetail, node)

		node.Self = false
	}

	fmt.Printf("Connected to peer %s\n", addrInfo.ID.Pretty())

	return nodeRef, nil
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

func (p *P2PService) NewP2PMessage(peerID peer.ID, protocolID protocol.ID, msg []byte) error {
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
