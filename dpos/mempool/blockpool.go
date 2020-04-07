// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package mempool

import (
	"errors"
	"github.com/elastos/Elastos.ELA/core/types/payload"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/core"
	dtype "github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"

	"sync"
)

const cachedCount = 6

type ConfirmInfo struct {
	Confirm *payload.Confirm
	Height  uint64
}

type BlockPool struct {
	Chain     *core.BlockChain
	IsCurrent func() bool

	sync.RWMutex
	blocks      map[common.Hash]*types.Block
	confirms    map[common.Hash]*payload.Confirm
}

func (bm *BlockPool) AppendConfirm(confirm *payload.Confirm) (bool,
	bool, error) {
	bm.Lock()
	defer bm.Unlock()

	return bm.appendConfirm(confirm)
}

func (bm *BlockPool) AddDposBlock(dposBlock *dtype.DposBlock) (bool, bool, error) {
	//// main version >=H1
	//if dposBlock.Block.Height >= bm.chainParams.CRCOnlyDPOSHeight {
	//	return bm.AppendDposBlock(dposBlock)
	//}
	//
	//return bm.Chain.ProcessBlock(dposBlock.Block, dposBlock.Confirm)
	return false, false, nil
}

func (bm *BlockPool) AppendDposBlock(dposBlock *dtype.DposBlock) (bool, bool, error) {
	bm.Lock()
	defer bm.Unlock()
	if !dposBlock.HaveConfirm {
		return bm.appendBlock(dposBlock)
	}
	return bm.appendBlockAndConfirm(dposBlock)
}

func (bm *BlockPool) appendBlock(dposBlock *dtype.DposBlock) (bool, bool, error) {
	// add block
	block := dposBlock.Block
	hash := block.Hash()
	if _, ok := bm.blocks[hash]; ok {
		return false, false, errors.New("duplicate block in pool")
	}
	//// verify block
	//if err := bm.Chain.CheckBlockSanity(block); err != nil {
	//	log.Info("[AppendBlock] check block sanity failed, ", err)
	//	return false, false, err
	//}
	//if block.Height == bm.Chain.GetHeight()+1 {
	//	prevNode, exist := bm.Chain.LookupNodeInIndex(&block.Header.Previous)
	//	if !exist {
	//		log.Info("[AppendBlock] check block context failed, there is no previous block on the chain")
	//		return false, false, errors.New("there is no previous block on the chain")
	//	}
	//	if err := bm.Chain.CheckBlockContext(block, prevNode); err != nil {
	//		log.Info("[AppendBlock] check block context failed, ", err)
	//		return false, false, err
	//	}
	//}
	//
	//bm.blocks[block.Hash()] = block
	//
	//// confirm block
	//inMainChain, isOrphan, err := bm.confirmBlock(hash)
	//if err != nil {
	//	log.Debug("[AppendDposBlock] ConfirmBlock failed, height", block.Height, "len(txs):",
	//		len(block.Transactions), "hash:", hash.String(), "err: ", err)
	//
	//	// Notify the caller that the new block without confirm was accepted.
	//	// The caller would typically want to react by relaying the inventory
	//	// to other peers.
	//	events.Notify(events.ETBlockAccepted, block)
	//	if block.Height == blockchain.DefaultLedger.Blockchain.GetHeight()+1 {
	//		events.Notify(events.ETNewBlockReceived, dposBlock)
	//	}
	//	return inMainChain, isOrphan, nil
	//}

	copyBlock := *dposBlock
	confirm := bm.confirms[hash]
	copyBlock.HaveConfirm = confirm != nil
	copyBlock.Confirm = confirm

	// notify new block received
	//events.Notify(events.ETNewBlockReceived, &copyBlock)

	//return inMainChain, isOrphan, nil
	return false, false, nil
}

func (bm *BlockPool) appendBlockAndConfirm(dposBlock *dtype.DposBlock) (bool, bool, error) {
	//block := dposBlock.Block
	//hash := block.Hash()
	//// verify block
	//if err := bm.Chain.CheckBlockSanity(block); err != nil {
	//	return false, false, err
	//}
	//// add block
	//bm.blocks[block.Hash()] = block
	//// confirm block
	//inMainChain, isOrphan, err := bm.appendConfirm(dposBlock.Confirm)
	//if err != nil {
	//	log.Debug("[appendBlockAndConfirm] ConfirmBlock failed, hash:", hash.String(), "err: ", err)
	//	return inMainChain, isOrphan, nil
	//}
	//
	//// notify new block received
	//events.Notify(events.ETNewBlockReceived, dposBlock)
	//
	//return inMainChain, isOrphan, nil
	return false, false, nil
}

func (bm *BlockPool) appendConfirm(confirm *payload.Confirm) (
	bool, bool, error) {

	// verify confirmation
	//if err := blockchain.ConfirmSanityCheck(confirm); err != nil {
	//	return false, false, err
	//}
	hash := common.Hash{}
	copy(hash[:], confirm.Proposal.BlockHash.Bytes()[:])
	bm.confirms[hash] = confirm

	inMainChain, isOrphan, err := bm.confirmBlock(hash)
	if err != nil {
		return inMainChain, isOrphan, err
	}
	//block := bm.blocks[confirm.Proposal.BlockHash]

	//// notify new confirm accepted.
	//events.Notify(events.ETConfirmAccepted, &ConfirmInfo{
	//	Confirm: confirm,
	//	Height:  block.NumberU64(),
	//})

	return inMainChain, isOrphan, nil
}

func (bm *BlockPool) ConfirmBlock(hash common.Hash) (bool, bool, error) {
	bm.Lock()
	inMainChain, isOrphan, err := bm.confirmBlock(hash)
	bm.Unlock()
	return inMainChain, isOrphan, err
}

func (bm *BlockPool) confirmBlock(hash common.Hash) (bool, bool, error) {
	log.Info("[ConfirmBlock] block hash:", hash)

	block, ok := bm.blocks[hash]
	if !ok {
		return false, false, errors.New("there is no block in pool when confirming block")
	}
	log.Info("[ConfirmBlock] block height:", block.NumberU64())

	//confirm, ok := bm.confirms[hash]
	//if !ok {
	//	return false, false, errors.New("there is no block confirmation in pool when confirming block")
	//}
	//
	//if !bm.Chain.BlockExists(&hash) {
	//	inMainChain, isOrphan, err := bm.Chain.ProcessBlock(block, confirm)
	//	if err != nil {
	//		return inMainChain, isOrphan, errors.New("add block failed," + err.Error())
	//	}
	//
	//	if !inMainChain && !isOrphan {
	//		if err := bm.CheckConfirmedBlockOnFork(bm.Chain.GetHeight(), block); err != nil {
	//			return inMainChain, isOrphan, err
	//		}
	//	}
	//
	//	if isOrphan && !inMainChain {
	//		bm.Chain.AddOrphanConfirm(confirm)
	//	}
	//
	//	if isOrphan || !inMainChain {
	//		return inMainChain, isOrphan, errors.New("add orphan block")
	//	}
	//} else {
	//	return false, false, errors.New("already processed block")
	//}

	return true, false, nil
}

func (bm *BlockPool) AddToBlockMap(block *types.Block) {
	bm.Lock()
	defer bm.Unlock()

	bm.blocks[block.Hash()] = block
}

func (bm *BlockPool) GetBlock(hash common.Hash) (*types.Block, bool) {
	bm.RLock()
	defer bm.RUnlock()

	block, ok := bm.blocks[hash]
	return block, ok
}

func (bm *BlockPool) GetDposBlockByHash(hash common.Hash) (*dtype.DposBlock, error) {
	bm.RLock()
	defer bm.RUnlock()

	if block := bm.blocks[hash]; block != nil {
		confirm := bm.confirms[hash]
		return &dtype.DposBlock{
			Block:       block,
			HaveConfirm: confirm != nil,
			Confirm:     confirm,
		}, nil
	}
	return nil, errors.New("not found dpos block")
}

func (bm *BlockPool) AddToConfirmMap(confirm *payload.Confirm) {
	bm.Lock()
	defer bm.Unlock()
	hash := common.Hash{}
	copy(hash[:], confirm.Proposal.BlockHash.Bytes()[:])
	bm.confirms[hash] = confirm
}

func (bm *BlockPool) GetConfirm(hash common.Hash) (
	*payload.Confirm, bool) {
	bm.Lock()
	defer bm.Unlock()

	confirm, ok := bm.confirms[hash]
	return confirm, ok
}

func (bm *BlockPool) CleanFinalConfirmedBlock(height uint64) {
	bm.Lock()
	defer bm.Unlock()

	for _, block := range bm.blocks {
		if block.NumberU64() < height-cachedCount {
			delete(bm.blocks, block.Hash())
			delete(bm.confirms, block.Hash())
		}
	}
}

func NewBlockPool() *BlockPool {
	return &BlockPool{
		blocks:      make(map[common.Hash]*types.Block),
		confirms:    make(map[common.Hash]*payload.Confirm),
	}
}
