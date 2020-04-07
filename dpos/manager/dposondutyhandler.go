// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package manager

import (
	"bytes"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/p2p/peer"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	com "github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/types/payload"
)

type DPOSOnDutyHandler struct {
	*DPOSHandlerSwitch
}

func (h *DPOSOnDutyHandler) ProcessAcceptVote(id peer.PID, p *payload.DPOSProposalVote) (succeed bool, finished bool) {
	log.Info("[Onduty-ProcessAcceptVote] start")
	defer log.Info("[Onduty-ProcessAcceptVote] end")

	currentProposal := h.proposalDispatcher.GetProcessingProposal()
	if currentProposal != nil && bytes.Equal(currentProposal.Hash().Bytes(), p.ProposalHash.Bytes()) && h.consensus.IsRunning() {
		log.Info("[OnVoteReceived] Received needed sign, collect it")
		return h.proposalDispatcher.ProcessVote(p, true)
	}

	return false, false
}

func (h *DPOSOnDutyHandler) ProcessRejectVote(id peer.PID, p *payload.DPOSProposalVote) (succeed bool, finished bool) {
	log.Info("[Onduty-ProcessRejectVote] start")

	currentProposal := h.proposalDispatcher.GetProcessingProposal()
	if currentProposal != nil && bytes.Equal(currentProposal.Hash().Bytes(), p.ProposalHash.Bytes()) && h.consensus.IsRunning() {
		return h.proposalDispatcher.ProcessVote(p, false)
	}

	return false, false
}

func (h *DPOSOnDutyHandler) ProcessProposal(id peer.PID, p *payload.DPOSProposal) (handled bool) {
	return false
}

func (h *DPOSOnDutyHandler) ChangeView(firstBlockHash *common.Hash) {

	if !h.tryCreateInactiveArbitratorsTx() {
		hash := com.Uint256{}
		copy(hash[:], firstBlockHash.Bytes()[:])
		b, ok := h.cfg.Manager.GetBlockCache().TryGetValue(hash)
		if !ok {
			log.Info("[OnViewChanged] get block failed for proposal")
		} else {
			log.Info("[OnViewChanged] start proposal")
			h.proposalDispatcher.CleanProposals(true)
			h.proposalDispatcher.StartProposal(b)
		}
	}
}

func (h *DPOSOnDutyHandler) TryStartNewConsensus(b *types.Block) bool {
	result := false

	if h.consensus.IsReady() {
		log.Info("[OnDuty][OnBlockReceived] received first unsigned block, start consensus")
		h.consensus.StartConsensus(b)
		h.proposalDispatcher.StartProposal(b)
		result = true
	} else { //finished
		log.Info("[OnDuty][OnBlockReceived] received unsigned block, record block")
		h.consensus.ProcessBlock(b)
		result = false
	}

	return result
}

func (h *DPOSOnDutyHandler) tryCreateInactiveArbitratorsTx() bool {
	//if h.proposalDispatcher.IsViewChangedTimeOut() {
	//	if h.cfg.Manager.isCRCArbiter() {
	//		tx, err := h.proposalDispatcher.CreateInactiveArbitrators()
	//		if err != nil {
	//			log.Warn("[tryCreateInactiveArbitratorsTx] create tx error: ", err)
	//			return false
	//		}
	//
	//		h.cfg.Network.BroadcastMessage(&msg.Tx{Serializable: tx})
	//	}
	//	return true
	//}
	return false
}
