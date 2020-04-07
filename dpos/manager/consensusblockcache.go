// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package manager

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	com "github.com/elastos/Elastos.ELA/common"
)

type ConsensusBlockCacheListener interface {
	OnBlockAdded(b *types.Block)
}

type ConsensusBlockCache struct {
	ConsensusBlocks    map[common.Hash]*types.Block
	ConsensusBlockList []common.Hash

	Listener ConsensusBlockCacheListener
}

func (c *ConsensusBlockCache) Reset(block *types.Block) {
	if block == nil {
		c.ConsensusBlocks = make(map[common.Hash]*types.Block)
		c.ConsensusBlockList = make([]common.Hash, 0)
		return
	}

	newConsensusBlocks := make(map[common.Hash]*types.Block)
	newConsensusBlockList := make([]common.Hash, 0)
	for _, b := range c.ConsensusBlocks {
		if b.Number().Cmp(block.Number()) < 0 {
			continue
		}
		newConsensusBlocks[b.Hash()] = b
		newConsensusBlockList = append(c.ConsensusBlockList, b.Hash())
	}
	c.ConsensusBlocks = newConsensusBlocks
	c.ConsensusBlockList = newConsensusBlockList
}

func (c *ConsensusBlockCache) AddValue(key common.Hash, value *types.Block) {
	c.ConsensusBlocks[key] = value
	c.ConsensusBlockList = append(c.ConsensusBlockList, key)

	if c.Listener != nil {
		c.Listener.OnBlockAdded(value)
	}
}

func (c *ConsensusBlockCache) TryGetValue(key com.Uint256) (*types.Block, bool) {
	hash := common.Hash{}
	copy(hash[:], key.Bytes()[:])
	value, ok := c.ConsensusBlocks[hash]

	return value, ok
}

func (c *ConsensusBlockCache) GetFirstArrivedBlockHash() (com.Uint256, bool) {
	if len(c.ConsensusBlockList) == 0 {
		return com.Uint256{}, false
	}
	hash := com.Uint256{}
	copy(hash[:], c.ConsensusBlockList[0].Bytes()[:])
	return hash, true
}
