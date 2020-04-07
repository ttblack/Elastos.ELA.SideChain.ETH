// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package types

import (
	"bytes"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"golang.org/x/crypto/sha3"
	"io"
)

type DPOSProposal struct {
	Sponsor    []byte
	BlockHash  common.Hash
	ViewOffset uint32

	// Signature values
	Sign   []byte

	hash *common.Hash
}

func (p *DPOSProposal) Data() []byte {
	buf := new(bytes.Buffer)

	if err := p.EncodeUnsigned(buf); err != nil {
		return []byte{0}
	}

	return buf.Bytes()
}

// EncodeRLP implements rlp.Encoder
func (p *DPOSProposal) EncodeUnsigned(w io.Writer) error {
	return rlp.Encode(w, &DPOSProposal{
		p.Sponsor,
		p.BlockHash,
		p.ViewOffset,
		nil,
		nil,
	})
}

// EncodeRLP implements rlp.Encoder
func (p *DPOSProposal) Serialize(w io.Writer) error {
	return rlp.Encode(w, p)
}

// DecodeRLP implements rlp.Decoder
func (p *DPOSProposal) Deserialize(s *rlp.Stream) error {
	err := s.Decode(p)
	if err != nil {
		log.Error("DPOSProposal Decode error:", err)
	}
	return err
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func (p *DPOSProposal) Hash() common.Hash {
	if (p.hash == nil || *p.hash == common.Hash{}) {
		hash := rlpHash(&DPOSProposal{
			p.Sponsor,
			p.BlockHash,
			p.ViewOffset,
			nil,
			nil,
		})
		p.hash = &hash
	}
	return *p.hash
}