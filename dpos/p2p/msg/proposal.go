// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"io"

	dtype "github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

const DefaultProposalMessageDataSize = 168 //67+32+4+65

type Proposal struct {
	Proposal dtype.DPOSProposal
}

func (m *Proposal) CMD() string {
	return CmdReceivedProposal
}

func (m *Proposal) MaxLength() uint32 {
	return DefaultProposalMessageDataSize
}

func (m *Proposal) Serialize(w io.Writer) error {
	return m.Proposal.Serialize(w)
}

func (m *Proposal) Deserialize(s *rlp.Stream) error {
	return m.Proposal.Deserialize(s)
}
