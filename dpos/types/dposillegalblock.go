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

type CoinType uint32

const (
	ELACoin = CoinType(0)

	IllegalBlockVersion byte = 0x00
)

type BlockEvidence struct {
	Header       []byte
	BlockConfirm []byte
	Signers      [][]byte

	hash *common.Hash
}

type DPOSIllegalBlocks struct {
	CoinType        CoinType
	BlockHeight     uint64
	Evidence        BlockEvidence
	CompareEvidence BlockEvidence

	hash *common.Hash
}

func (b *BlockEvidence) SerializeUnsigned(w io.Writer) error {
	return nil
}

func (b *BlockEvidence) SerializeOthers(w io.Writer) error {

	return nil
}

func (b *BlockEvidence) Serialize(w io.Writer) error {
	return nil
}

func (b *BlockEvidence) DeserializeUnsigned(r io.Reader) error {

	return nil
}

func (b *BlockEvidence) DeserializeOthers(r io.Reader) (err error) {
	return nil
}

func (b *BlockEvidence) Deserialize(r io.Reader) error {
	return nil
}

func (b *BlockEvidence) BlockHash() common.Hash {
	if b.hash == nil {
	 	hash := rlpHash(b)
		b.hash = &hash
	}
	return *b.hash
}

func (d *DPOSIllegalBlocks) Data(version byte) []byte {
	buf := new(bytes.Buffer)
	if err := d.Serialize(buf, version); err != nil {
		return []byte{0}
	}
	return buf.Bytes()
}

func (d *DPOSIllegalBlocks) SerializeUnsigned(w io.Writer, version byte) error {


	return nil
}

func (d *DPOSIllegalBlocks) Serialize(w io.Writer, version byte) error {
	return nil
}

func (d *DPOSIllegalBlocks) DeserializeUnsigned(r io.Reader, version byte) error {
	return nil
}

func (d *DPOSIllegalBlocks) Deserialize(r io.Reader, version byte) error {
	return nil
}

func (d *DPOSIllegalBlocks) Hash() common.Hash {
	if d.hash == nil {
		hash := rlpHash(d)
		d.hash = &hash
	}
	return *d.hash
}

func (d *DPOSIllegalBlocks) GetBlockHeight() uint64 {
	return d.BlockHeight
}

func (d *DPOSIllegalBlocks) Type() IllegalDataType {
	return IllegalBlock
}
