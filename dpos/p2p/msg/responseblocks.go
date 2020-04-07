// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
)

//todo move to config
const DefaultResponseBlocksMessageDataSize = 8000000 * 10

type ResponseBlocks struct {
	Command       string
	BlockConfirms []*types.DposBlock
}

func (m *ResponseBlocks) CMD() string {
	return CmdResponseBlocks
}

func (m *ResponseBlocks) MaxLength() uint32 {
	return DefaultResponseBlocksMessageDataSize
}

func (m *ResponseBlocks) Serialize(w io.Writer) error {
	return nil
}

func (m *ResponseBlocks) Deserialize(s *rlp.Stream) error {
	return nil
}
