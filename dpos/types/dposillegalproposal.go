// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package types

import (
	"bytes"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
)

const (
	IllegalProposalVersion byte = 0x00
)

type ProposalEvidence struct {
	Proposal    DPOSProposal
	BlockHeader []byte
	BlockHeight uint64
}

type DPOSIllegalProposals struct {
	Evidence        ProposalEvidence
	CompareEvidence ProposalEvidence

	hash *common.Hash
}

func (d *ProposalEvidence) Serialize(w io.Writer) error {
	return nil
}

func (d *ProposalEvidence) Deserialize(r io.Reader) (err error) {
	return nil
}

func (d *DPOSIllegalProposals) Data(version byte) []byte {
	buf := new(bytes.Buffer)
	if err := d.Serialize(buf, version); err != nil {
		return []byte{0}
	}
	return buf.Bytes()
}

func (d *DPOSIllegalProposals) Serialize(w io.Writer, version byte) error {
	return nil
}

func (d *DPOSIllegalProposals) Deserialize(r io.Reader, version byte) error {
	return nil
}

func (d *DPOSIllegalProposals) Hash() common.Hash {
	if d.hash == nil {
		hash := rlpHash(d)
		d.hash = &hash
	}
	return *d.hash
}

func (d *DPOSIllegalProposals) GetBlockHeight() uint64 {
	return d.Evidence.BlockHeight
}

func (d *DPOSIllegalProposals) Type() IllegalDataType {
	return IllegalProposal
}
