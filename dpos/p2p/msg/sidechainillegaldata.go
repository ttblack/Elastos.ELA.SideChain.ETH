// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/elanet/pact"
)

type SidechainIllegalData struct {
	Data types.SidechainIllegalData
}

func (msg *SidechainIllegalData) CMD() string {
	return CmdSidechainIllegalData
}

func (msg *SidechainIllegalData) MaxLength() uint32 {
	return pact.MaxBlockSize
}

func (msg *SidechainIllegalData) Serialize(w io.Writer) error {
	return nil
}

func (msg *SidechainIllegalData) Deserialize(s *rlp.Stream) error {
	return nil
}
