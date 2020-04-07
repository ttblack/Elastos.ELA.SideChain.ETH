// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
)

type RequestProposal struct {
	ProposalHash common.Hash
}

func (msg *RequestProposal) CMD() string {
	return CmdRequestProposal
}

func (msg *RequestProposal) MaxLength() uint32 {
	return 32
}

func (msg *RequestProposal) Serialize(w io.Writer) error {

	return nil
}

func (msg *RequestProposal) Deserialize(s *rlp.Stream) error {
	return nil
}
