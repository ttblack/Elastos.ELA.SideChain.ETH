// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	dtype "github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"
)

const MaxIllegalVoteSize = 1000000

type IllegalVotes struct {
	Votes dtype.DPOSIllegalVotes
}

func (msg *IllegalVotes) CMD() string {
	return CmdIllegalVotes
}

func (msg *IllegalVotes) MaxLength() uint32 {
	return MaxIllegalVoteSize
}

func (msg *IllegalVotes) Serialize(w io.Writer) error {
	if err := msg.Votes.Serialize(w, dtype.IllegalVoteVersion); err != nil {
		return err
	}

	return nil
}

func (msg *IllegalVotes) Deserialize(s *rlp.Stream) error {
	if err := msg.Votes.Deserialize(s, dtype.IllegalVoteVersion); err != nil {
		return err
	}

	return nil
}
