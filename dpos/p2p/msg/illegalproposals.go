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

const MaxIllegalProposalSize = 1000000

type IllegalProposals struct {
	Proposals dtype.DPOSIllegalProposals
}

func (msg *IllegalProposals) CMD() string {
	return CmdIllegalProposals
}

func (msg *IllegalProposals) MaxLength() uint32 {
	return MaxIllegalProposalSize
}

func (msg *IllegalProposals) Serialize(w io.Writer) error {
	if err := msg.Proposals.Serialize(w,
		dtype.IllegalProposalVersion); err != nil {
		return err
	}

	return nil
}

func (msg *IllegalProposals) Deserialize(s *rlp.Stream) error {
	//if err := msg.Proposals.Deserialize(r,
	//	dtype.IllegalProposalVersion); err != nil {
	//	return err
	//}

	return nil
}
