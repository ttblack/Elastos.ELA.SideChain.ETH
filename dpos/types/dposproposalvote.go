// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package types

import (
	"bytes"
	com "github.com/elastos/Elastos.ELA/common"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"

)

type DPOSProposalVote struct {
	ProposalHash com.Uint256

	Signer []byte
	Accept bool
	Sign   []byte

	hash *common.Hash
}

func (v *DPOSProposalVote) Data() []byte {
	buf := new(bytes.Buffer)
	if err := v.SerializeUnsigned(buf); err != nil {
		return []byte{0}
	}

	return buf.Bytes()
}

func (v *DPOSProposalVote) SerializeUnsigned(w io.Writer) error {
	return rlp.Encode(w, &DPOSProposalVote{
		v.ProposalHash,
		v.Signer,
		v.Accept,
		nil,
		nil,
	})
}

func (v *DPOSProposalVote) Serialize(w io.Writer) error {
	return rlp.Encode(w, v)
}

func (v *DPOSProposalVote) Deserialize(s *rlp.Stream) error {
	err := s.Decode(v)
	if err != nil {
		log.Error("DPOSProposalVote Decode error:", err)
	}

	return err
}

func (v *DPOSProposalVote) Hash() common.Hash {
	if (v.hash == nil || *v.hash == common.Hash{}) {
		hash := rlpHash(&DPOSProposalVote{
			v.ProposalHash,
			v.Signer,
			v.Accept,
			nil,
			nil,
		})
		v.hash = &hash
	}
	return *v.hash
}
