// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dposp2p"
)

// Ensure GetBlockByHash implement p2p.Message interface.
var _ dposp2p.Message = (*GetBlock)(nil)

type GetBlock struct {
	Inventory
}

func NewGetBlock(blockHash common.Hash) *GetBlock {
	getBlock := new(GetBlock)
	getBlock.BlockHash = blockHash
	return getBlock
}

func (msg *GetBlock) CMD() string {
	return CmdGetBlock
}
