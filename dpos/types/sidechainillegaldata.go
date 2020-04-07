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

const SidechainIllegalDataVersion byte = 0x00

type SidechainIllegalEvidence struct {
	DataHash common.Hash
}

type SidechainIllegalData struct {
	IllegalType         IllegalDataType
	Height              uint32
	IllegalSigner       []byte
	Evidence            SidechainIllegalEvidence
	CompareEvidence     SidechainIllegalEvidence
	GenesisBlockAddress string
	Signs               [][]byte

	hash *common.Hash
}

func (s *SidechainIllegalEvidence) Serialize(w io.Writer) error {
	return nil
}

func (s *SidechainIllegalEvidence) Deserialize(r io.Reader) error {
	return nil
}

func (s *SidechainIllegalData) Data(version byte) []byte {
	buf := new(bytes.Buffer)
	if err := s.Serialize(buf, version); err != nil {
		return []byte{0}
	}
	return buf.Bytes()
}

func (s *SidechainIllegalData) Type() IllegalDataType {
	return s.IllegalType
}

func (s *SidechainIllegalData) GetBlockHeight() uint32 {
	return s.Height
}

func (s *SidechainIllegalData) SerializeUnsigned(w io.Writer, version byte) error {
	return nil
}

func (s *SidechainIllegalData) Serialize(w io.Writer, version byte) error {
	return nil
}

func (s *SidechainIllegalData) DeserializeUnsigned(r io.Reader,
	version byte) error {
	var err error
	return err
}

func (s *SidechainIllegalData) Deserialize(r io.Reader, version byte) error {
	var err error
	if err = s.DeserializeUnsigned(r, version); err != nil {
		return err
	}
	return nil
}

func (s *SidechainIllegalData) Hash() common.Hash {
	if s.hash == nil {
		hash := rlpHash(s)
		s.hash = &hash
	}
	return *s.hash
}
