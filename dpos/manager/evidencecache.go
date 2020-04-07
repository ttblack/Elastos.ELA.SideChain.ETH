// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package manager

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	com "github.com/elastos/Elastos.ELA/common"
	"github.com/elastos/Elastos.ELA/core/types/payload"
)

type evidenceCache struct {
	evidences map[com.Uint256]payload.DPOSIllegalData
}

func (e *evidenceCache) AddEvidence(evidence payload.DPOSIllegalData) {
	if evidence != nil {
		e.evidences[evidence.Hash()] = evidence
	}
}

func (e *evidenceCache) IsBlockValid(block *types.Block) bool {
	if len(e.evidences) == 0 {
		return true
	}

	necessaryEvidences := make(map[com.Uint256]interface{})
	for k, v := range e.evidences {
		tolerance := WaitHeightTolerance
		if v.Type() == payload.IllegalBlock ||
			v.Type() == payload.InactiveArbitrator {
			tolerance = 0
		}
		if uint64(v.GetBlockHeight() + tolerance) <= block.NumberU64() {
			necessaryEvidences[k] = nil
		}
	}

	for _, t := range block.Transactions() {
		if hash, ok := e.tryGetEvidenceHash(t); ok {
			if _, hasEvidence := necessaryEvidences[hash]; hasEvidence {
				delete(necessaryEvidences, hash)
			}
		}
	}

	log.Debug("[IsBlockValid] necessaryEvidences count left count :",
		len(necessaryEvidences))

	return len(necessaryEvidences) == 0
}

func (e *evidenceCache) Reset(block *types.Block) {
	for _, t := range block.Transactions() {
		if hash, ok := e.tryGetEvidenceHash(t); ok {
			if _, hasEvidence := e.evidences[hash]; hasEvidence {
				delete(e.evidences, hash)
			}
		}
	}
}

func (e *evidenceCache) TryDelete(hash com.Uint256) {
	if _, hasEvidence := e.evidences[hash]; hasEvidence {
		delete(e.evidences, hash)
	}
}

func (e *evidenceCache) tryGetEvidenceHash(tx *types.Transaction) (com.Uint256, bool) {
	var hash com.Uint256
	result := true

	//switch tx.TxType {
	//case types.IllegalProposalEvidence:
	//	proposalPayload := tx.Payload.(*payload.DPOSIllegalProposals)
	//	hash = proposalPayload.Hash()
	//case types.IllegalVoteEvidence:
	//	votePayload := tx.Payload.(*payload.DPOSIllegalVotes)
	//	hash = votePayload.Hash()
	//case types.IllegalBlockEvidence:
	//	blockPayload := tx.Payload.(*payload.DPOSIllegalBlocks)
	//	hash = blockPayload.Hash()
	//case types.IllegalSidechainEvidence:
	//	sidechainPayload := tx.Payload.(*payload.SidechainIllegalData)
	//	hash = sidechainPayload.Hash()
	//case types.InactiveArbitrators:
	//	inactiveArbitrators := tx.Payload.(*payload.InactiveArbitrators)
	//	hash = inactiveArbitrators.Hash()
	//default:
	//	result = false
	//}

	return hash, result
}
