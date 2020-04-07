package elanet

import (
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/mempool"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/elanet/pact"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/elanet/peer"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/elanet/routes"
	svr "github.com/elastos/Elastos.ELA/p2p/server"

	"github.com/elastos/Elastos.ELA/p2p"
	"github.com/elastos/Elastos.ELA/p2p/msg"
)
const (
	// defaultServices describes the default services that are supported by
	// the server.
	defaultServices = pact.SFNodeNetwork | pact.SFTxFiltering | pact.SFNodeBloom

	// maxNonNodePeers defines the maximum count of accepting non-node peers.
	maxNonNodePeers = 100
)

// server provides a server for handling communications to and from peers.
type server struct {
	svr.IServer

	blockMemPool *mempool.BlockPool
	routes       *routes.Routes

	peerQueue    chan interface{}
	relayInv     chan relayMsg
	quit         chan struct{}
	services     pact.ServiceFlag
}

// newPeerMsg represent a new connected peer.
type newPeerMsg struct {
	svr.IPeer
	reply chan struct{}
}

// donePeerMsg represent a disconnected peer.
type donePeerMsg struct {
	svr.IPeer
	reply chan struct{}
}

// relayMsg packages an inventory vector along with the newly discovered
// inventory so the relay has access to that information.
type relayMsg struct {
	invVect *msg.InvVect
	data    interface{}
}

// serverPeer extends the peer to maintain state shared by the server and
// the blockmanager.
type serverPeer struct {
	*peer.Peer
	server        *server
	syncManager   *netsync.SyncManager
	isWhitelisted bool
	//filter        *filter.Filter
	quit          chan struct{}
	// The following chans are used to sync blockmanager and server.
	txProcessed    chan struct{}
	blockProcessed chan struct{}
}

// newServerPeer returns a new serverPeer instance. The peer needs to be set by
// the caller.
func newServerPeer(s *server) *serverPeer {
	return &serverPeer{
		server:         s,
	}
}

// Services returns the service flags the server supports.
func (s *server) Services() pact.ServiceFlag {
	return s.services
}

// NewServer returns a new elanet server configured to listen on addr for the
// network type specified by chainParams.  Use start to begin accepting
// connections from peers.
func NewServer(dataDir string, cfg *Config) (*server, error) {
	services := defaultServices
	params := cfg.ChainParams
	if params.DisableTxFilters {
		services &^= pact.SFNodeBloom
		services &^= pact.SFTxFiltering
	}

	// If no listeners added, create default listener.
	listeneAddr := []string{fmt.Sprint(":", params.DPoSPort)}

	svrCfg := svr.NewDefaultConfig(
		params.Magic, pact.DPOSStartVersion, uint64(services),
		params.DPoSPort, params.DNSSeeds, listeneAddr,
		nil, nil, makeEmptyMessage,
		func() uint64 { return cfg.Chain.CurrentHeader().Number.Uint64() },
	)
	svrCfg.DataDir = dataDir
	svrCfg.PermanentPeers = params.PermanentPeers

	s := server{
		blockMemPool: cfg.BlockMemPool,
		routes:       cfg.Routes,
	}
	svrCfg.OnNewPeer = s.NewPeer
	svrCfg.OnDonePeer = s.DonePeer
	//
	p2pServer, err := svr.NewServer(svrCfg)
	if err != nil {
		return nil, err
	}
	s.IServer = p2pServer
	return &s, nil
}

// NewPeer adds a new peer that has already been connected to the server.
func (s *server) NewPeer(p svr.IPeer) {
	reply := make(chan struct{})
	s.peerQueue <- newPeerMsg{p, reply}
	<-reply
}

// DonePeer removes a peer that has already been connected to the server by ip.
func (s *server) DonePeer(p svr.IPeer) {
	reply := make(chan struct{})
	s.peerQueue <- donePeerMsg{p, reply}
	<-reply
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (s *server) RelayInventory(invVect *msg.InvVect, data interface{}) {
	s.relayInv <- relayMsg{invVect: invVect, data: data}
}

func (s *server) IsCurrent() bool {
	return s.syncManager.IsCurrent()
}

func makeEmptyMessage(cmd string) (p2p.Message, error) {
	var message p2p.Message
	switch cmd {
	case p2p.CmdMemPool:
		message = &msg.MemPool{}

	case p2p.CmdTx:

	case p2p.CmdBlock:
		//message = msg.NewBlock(&types.DposBlock{})

	case p2p.CmdInv:
		message = &msg.Inv{}

	case p2p.CmdNotFound:
		message = &msg.NotFound{}

	case p2p.CmdGetData:
		message = &msg.GetData{}

	case p2p.CmdGetBlocks:
		message = &msg.GetBlocks{}

	case p2p.CmdFilterAdd:
		message = &msg.FilterAdd{}

	case p2p.CmdFilterClear:
		message = &msg.FilterClear{}

	case p2p.CmdFilterLoad:
		message = &msg.FilterLoad{}

	case p2p.CmdTxFilter:
		message = &msg.TxFilterLoad{}

	case p2p.CmdReject:
		message = &msg.Reject{}

	case p2p.CmdDAddr:
		message = &msg.DAddr{}

	default:
		return nil, fmt.Errorf("unhandled command [%s]", cmd)
	}
	return message, nil
}
