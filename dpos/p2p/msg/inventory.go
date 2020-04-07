// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"
)

type Inventory struct {
	BlockHash common.Hash
}

func NewInventory(blockHash common.Hash) *Inventory {
	inv := new(Inventory)
	inv.BlockHash = blockHash
	return inv
}

func (msg *Inventory) CMD() string {
	return CmdInv
}

func (msg *Inventory) MaxLength() uint32 {
	return 32
}

func (msg *Inventory) Serialize(writer io.Writer) error {
	return nil
}

func (msg *Inventory) Deserialize(s *rlp.Stream) error {
	return nil
}
