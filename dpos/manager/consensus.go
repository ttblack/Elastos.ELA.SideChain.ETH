// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package manager

import (
	"bytes"
	"github.com/elastos/Elastos.ELA/dpos/p2p/msg"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/log"
	elog "github.com/elastos/Elastos.ELA.SideChain.ETH/log"
)

const (
	consensusReady = iota
	consensusRunning
)

type Consensus struct {
	consensusStatus uint32
	viewOffset      uint32

	manager     *DPOSManager
	currentView view
}

func NewConsensus(manager *DPOSManager, tolerance time.Duration,
	viewListener ViewListener) *Consensus {
	c := &Consensus{
		consensusStatus: consensusReady,
		viewOffset:      0,
		manager:         manager,
		currentView: view{
			account:     	manager.account,
			signTolerance: tolerance,
			listener:      viewListener,
			arbitrators:   manager.arbitrators,
		},
	}

	return c
}

func (c *Consensus) IsOnDuty() bool {
	return c.currentView.IsOnDuty()
}

func (c *Consensus) SetOnDuty(onDuty bool) {
	c.currentView.SetOnDuty(onDuty)
}

func (c *Consensus) SetRunning() {
	c.consensusStatus = consensusRunning
	c.resetViewOffset()
}

func (c *Consensus) SetReady() {
	c.consensusStatus = consensusReady
	c.resetViewOffset()
}

func (c *Consensus) IsRunning() bool {
	return c.consensusStatus == consensusRunning
}

func (c *Consensus) IsReady() bool {
	return c.consensusStatus == consensusReady
}

func (c *Consensus) IsArbitratorOnDuty(arbitrator []byte) bool {
	return bytes.Equal(c.GetOnDutyArbitrator(), arbitrator)
}

func (c *Consensus) GetOnDutyArbitrator() []byte {
	return c.manager.GetArbitrators().GetNextOnDutyArbitrator(c.viewOffset)
}

func (c *Consensus) StartConsensus(b *types.Block) {
	elog.Info("[StartConsensus] consensus start")
	defer elog.Info("[StartConsensus] consensus end")

	now := c.manager.timeSource.AdjustedTime()
	c.manager.GetBlockCache().Reset(b)
	c.SetRunning()

	c.manager.GetBlockCache().AddValue(b.Hash(), b)
	c.currentView.ResetView(now)
	viewEvent := log.ViewEvent{
		OnDutyArbitrator: common.Bytes2Hex(c.GetOnDutyArbitrator()),
		StartTime:        now,
		Offset:           c.GetViewOffset(),
		Height:           b.Number().Uint64(),
	}
	c.manager.dispatcher.cfg.EventMonitor.OnViewStarted(&viewEvent)
}

func (c *Consensus) GetViewOffset() uint32 {
	return c.viewOffset
}

func (c *Consensus) ProcessBlock(b *types.Block) {
	c.manager.GetBlockCache().AddValue(b.Hash(), b)
}

func (c *Consensus) ChangeView() {
	c.currentView.ChangeView(&c.viewOffset, c.manager.timeSource.AdjustedTime())
}

func (c *Consensus) TryChangeView() bool {
	if c.IsRunning() {
		return c.currentView.TryChangeView(&c.viewOffset, c.manager.timeSource.AdjustedTime())
	}
	return false
}

func (c *Consensus) CollectConsensusStatus(status *msg.ConsensusStatus) error {
	status.ConsensusStatus = c.consensusStatus
	status.ViewOffset = c.viewOffset
	status.ViewStartTime = c.currentView.GetViewStartTime()
	elog.Info("[CollectConsensusStatus] status.ConsensusStatus:", status.ConsensusStatus)
	return nil
}

func (c *Consensus) RecoverFromConsensusStatus(status *msg.ConsensusStatus) error {
	elog.Info("[RecoverFromConsensusStatus] status.ConsensusStatus:", status.ConsensusStatus)
	c.consensusStatus = status.ConsensusStatus
	c.viewOffset = status.ViewOffset
	c.currentView.ResetView(status.ViewStartTime)
	return nil
}

func (c *Consensus) resetViewOffset() {
	c.viewOffset = 0
}
