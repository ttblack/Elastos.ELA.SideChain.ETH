// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

const DefaultVoteMessageDataSize = 297 //164+67+1+65

type Vote struct {
	Command string
	Vote    types.DPOSProposalVote
}

func (msg *Vote) CMD() string {
	return msg.Command
}

func (msg *Vote) MaxLength() uint32 {
	return DefaultVoteMessageDataSize
}

func (msg *Vote) Serialize(w io.Writer) error {
	return msg.Vote.Serialize(w)
}

func (msg *Vote) Deserialize(s *rlp.Stream) error {
	return msg.Vote.Deserialize(s)
}