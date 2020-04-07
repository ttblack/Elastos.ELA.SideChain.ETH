// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package types

import (
	"bytes"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
)

const (
	IllegalVoteVersion byte = 0x00
)

type VoteEvidence struct {
	ProposalEvidence
	Vote DPOSProposalVote
}

type DPOSIllegalVotes struct {
	Evidence        VoteEvidence
	CompareEvidence VoteEvidence

	hash *common.Hash
}

func (d *VoteEvidence) Serialize(w io.Writer) error {
	if err := d.Vote.Serialize(w); err != nil {
		return err
	}

	if err := d.ProposalEvidence.Serialize(w); err != nil {
		return err
	}

	return nil
}

func (d *VoteEvidence) Deserialize(s *rlp.Stream) (err error) {
	if err := d.Vote.Deserialize(s); err != nil {
		return err
	}

	//if err := d.ProposalEvidence.Deserialize(s); err != nil {
	//	return err
	//}

	return nil
}

func (d *DPOSIllegalVotes) Data(version byte) []byte {
	buf := new(bytes.Buffer)
	if err := d.Serialize(buf, version); err != nil {
		return []byte{0}
	}
	return buf.Bytes()
}

func (d *DPOSIllegalVotes) Serialize(w io.Writer, version byte) error {
	if err := d.Evidence.Serialize(w); err != nil {
		return err
	}

	if err := d.CompareEvidence.Serialize(w); err != nil {
		return err
	}

	return nil
}

func (d *DPOSIllegalVotes) Deserialize(s *rlp.Stream, version byte) error {
	if err := d.Evidence.Deserialize(s); err != nil {
		return err
	}

	if err := d.CompareEvidence.Deserialize(s); err != nil {
		return err
	}

	return nil
}

func (d *DPOSIllegalVotes) Hash() common.Hash {
	if d.hash == nil {
		hash := rlpHash(d)
		d.hash = &hash
	}
	return *d.hash
}

func (d *DPOSIllegalVotes) GetBlockHeight() uint64 {
	return d.Evidence.BlockHeight
}

func (d *DPOSIllegalVotes) Type() IllegalDataType {
	return IllegalVote
}
