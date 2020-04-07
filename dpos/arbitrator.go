// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package dpos

import (
	"bytes"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/p2p"
	"github.com/elastos/Elastos.ELA/core/types/payload"
	ep2p "github.com/elastos/Elastos.ELA/p2p"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/accounts"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/dtime"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/events"
	dlog "github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/log"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/manager"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/mempool"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/p2p/peer"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/state"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/store"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/elanet"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/params"
)

type Config struct {
	EnableEventLog    bool
	EnableEventRecord bool
	Arbitrators       state.Arbitrators
	Store             store.IDposStore
	Server            elanet.Server
	TxMemPool         *core.TxPool
	BlockMemPool      *mempool.BlockPool
	ChainParams       *params.DposConfig
	Broadcast         func(msg ep2p.Message)
	AnnounceAddr      func()
}

type Arbitrator struct {
	cfg            Config
	account        accounts.Wallet
	enableViewLoop bool
	network        *network
	dposManager    *manager.DPOSManager
}

func (a *Arbitrator) Start() {
	a.network.Start()

	go a.changeViewLoop()
	//go a.recover()
}

func (a *Arbitrator) recover() {
	for {
		if a.cfg.Server.IsCurrent() && a.dposManager.GetArbitrators().
			HasArbitersMinorityCount(len(a.network.GetActivePeers())) {
			a.network.recoverChan <- true
			return
		}
		time.Sleep(time.Second)
	}
}

func (a *Arbitrator) Stop() error {
	a.enableViewLoop = false

	if err := a.network.Stop(); err != nil {
		return err
	}

	return nil
}

func (a *Arbitrator) GetArbiterPeersInfo() []*p2p.PeerInfo {
	return a.network.p2pServer.DumpPeersInfo()
}

func (a *Arbitrator) OnIllegalBlockTxReceived(p *payload.DPOSIllegalBlocks) {
	log.Info("[OnIllegalBlockTxReceived] listener received illegal block tx")
	if p.CoinType != payload.ELACoin {
		a.network.PostIllegalBlocksTask(p)
	}
}

func (a *Arbitrator) OnInactiveArbitratorsTxReceived(
	p *payload.InactiveArbitrators) {
	log.Info("[OnInactiveArbitratorsTxReceived] listener received " +
		"inactive arbitrators tx")

	if !a.cfg.Arbitrators.IsArbitrator(a.dposManager.GetAccount()) {
		//isEmergencyCandidate := false

		candidates := a.cfg.Arbitrators.GetCandidates()
		inactiveEliminateCount := a.cfg.Arbitrators.GetArbitersCount() / 3
		for i := 0; i < len(candidates) && i < inactiveEliminateCount; i++ {
			if bytes.Equal(candidates[i], a.dposManager.GetAccount().Bytes()) {
				//isEmergencyCandidate = true
			}
		}

		//if isEmergencyCandidate {
		//	if err := a.cfg.Arbitrators.ProcessSpecialTxPayload(p,
		//		blockchain.DefaultLedger.Blockchain.GetHeight()); err != nil {
		//		log.Error("[OnInactiveArbitratorsTxReceived] force change "+
		//			"arbitrators error: ", err)
		//		return
		//	}
		//	go a.recover()
		//}
	} else {
		a.network.PostInactiveArbitersTask(p)
	}
}

func (a *Arbitrator) OnSidechainIllegalEvidenceReceived(
	data *payload.SidechainIllegalData) {
	log.Info("[OnSidechainIllegalEvidenceReceived] listener received" +
		" sidechain illegal evidence")
	a.network.PostSidechainIllegalDataTask(data)
}

func (a *Arbitrator) OnBlockReceived(b *types.Block, confirmed bool) {
	if !a.cfg.Server.IsCurrent() {
		return
	}
	log.Info("[OnBlockReceived] listener received block")
	a.network.PostBlockReceivedTask(b, confirmed)
}

func (a *Arbitrator) OnConfirmReceived(p *mempool.ConfirmInfo) {
	//if !a.cfg.Server.IsCurrent() {
	//	return
	//}
	//log.Info("[OnConfirmReceived] listener received confirm")
	//a.network.PostConfirmReceivedTask(p)
}

func (a *Arbitrator) OnPeersChanged(peers []peer.PID) {
	a.network.UpdatePeers(peers)
}

func (a *Arbitrator) changeViewLoop() {
	for a.enableViewLoop {
		a.network.PostChangeViewTask()

		time.Sleep(1 * time.Second)
	}
}

// OnCipherAddr will be invoked when an address cipher received.
func (a *Arbitrator) OnCipherAddr(pid peer.PID, cipher []byte) {
	//addr, err := a.account.DecryptAddr(cipher)
	//if err != nil {
	//	log.Error("decrypt address cipher error %s", err)
	//	return
	//}
	//a.network.p2pServer.AddAddr(pid, addr)
}

func NewArbitrator(cfg Config, account accounts.Wallet) (*Arbitrator) {
	medianTime := dtime.NewMedianTime()
	dposManager := manager.NewManager(manager.DPOSManagerConfig{
		Account: 	 account.Accounts()[0].Address,
		Arbitrators: cfg.Arbitrators,
		TimeSource:  medianTime,
		Server:      cfg.Server,
	})

	var network, err = NewDposNetwork(NetworkConfig{
		ChainParams: cfg.ChainParams,
		Account:     account,
		MedianTime:  medianTime,
		Listener:    dposManager,
	})
	if err != nil {
		log.Error("Init p2p network error")
		return nil
	}

	eventMonitor := dlog.NewEventMonitor()

	if cfg.EnableEventLog {
		eventLogs := &dlog.EventLogs{}
		eventMonitor.RegisterListener(eventLogs)
	}

	if cfg.EnableEventRecord {
		eventRecorder := &store.EventRecord{}
		eventRecorder.Initialize(cfg.Store)
		eventMonitor.RegisterListener(eventRecorder)
	}
	dposHandlerSwitch := manager.NewHandler(manager.DPOSHandlerConfig{
		Network:     network,
		Manager:     dposManager,
		Monitor:     eventMonitor,
		Arbitrators: cfg.Arbitrators,
		TimeSource:  medianTime,
	})
	consensus := manager.NewConsensus(dposManager, cfg.ChainParams.SignTolerance, dposHandlerSwitch)
	proposalDispatcher, illegalMonitor := manager.NewDispatcherAndIllegalMonitor(
		manager.ProposalDispatcherConfig{
			EventMonitor: eventMonitor,
			Consensus:    consensus,
			Network:      network,
			Manager:      dposManager,
			Account:      account,
			ChainParams:  cfg.ChainParams,
			TimeSource:   medianTime,
			EventStoreAnalyzerConfig: store.EventStoreAnalyzerConfig{
				Store:       cfg.Store,
				Arbitrators: cfg.Arbitrators,
			},
		})
	dposHandlerSwitch.Initialize(proposalDispatcher, consensus)

	dposManager.Initialize(dposHandlerSwitch, proposalDispatcher, consensus,
		network, illegalMonitor, cfg.BlockMemPool, cfg.TxMemPool, cfg.Broadcast)
	network.Initialize(manager.DPOSNetworkConfig{
		ProposalDispatcher: proposalDispatcher,
		Store:              cfg.Store,
		Account:           	account.Accounts()[0].Address,
		AnnounceAddr:       cfg.AnnounceAddr,
	})

	cfg.Store.StartEventRecord()

	a := Arbitrator{
		cfg:            cfg,
		account:        account,
		enableViewLoop: true,
		dposManager:    dposManager,
		network:        network,
	}

	events.Subscribe(func(e *events.Event) {
		switch e.Type {
		case events.ETNewBlockReceived:
			//block := e.Data.(*dtypes.DposBlock)
			//go a.OnBlockReceived(block.Block, block.HaveConfirm)

		case events.ETConfirmAccepted:
			//go a.OnConfirmReceived(e.Data.(*mempool.ConfirmInfo))

		case events.ETDirectPeersChanged:
			//a.OnPeersChanged(e.Data.([]peer.PID))

		case events.ETTransactionAccepted:
			//tx := e.Data.(*types.Transaction)
			//if tx.IsIllegalBlockTx() {
			//	a.OnIllegalBlockTxReceived(tx.Payload.(*payload.DPOSIllegalBlocks))
			//} else if tx.IsInactiveArbitrators() {
			//	a.OnInactiveArbitratorsTxReceived(tx.Payload.(*payload.InactiveArbitrators))
			//}
		}
	})

	return &a
}

func (a *Arbitrator) IsArbiter() bool {
	return a.dposManager.GetArbitrators().IsArbitrator(a.dposManager.GetAccount())
}

func (a *Arbitrator) IsOnDuty() bool {
	selfAccount := a.dposManager.GetAccount()
	account := a.dposManager.GetArbitrators().GetOnDutyArbitrator()
	return bytes.Equal(selfAccount.Bytes(), account)
}

func (a *Arbitrator) GetProducerCount() int {
	return a.dposManager.GetArbitrators().GetArbitersCount()
}
