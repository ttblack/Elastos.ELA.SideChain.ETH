// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"io"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"

)

type ConsensusStatus struct {
	ConsensusStatus uint32
	ViewOffset      uint32
	ViewStartTime   time.Time

	AcceptVotes      []types.DPOSProposalVote
	RejectedVotes    []types.DPOSProposalVote
	PendingProposals []types.DPOSProposal
	PendingVotes     []types.DPOSProposalVote
}

func (s *ConsensusStatus) Serialize(w io.Writer) error {
	return nil
}

func (s *ConsensusStatus) Deserialize(stream *rlp.Stream) error {
	return nil
}
