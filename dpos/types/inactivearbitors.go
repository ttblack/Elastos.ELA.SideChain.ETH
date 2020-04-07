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

const InactiveArbitratorsVersion byte = 0x00

type InactiveArbitrators struct {
	Sponsor     []byte
	Arbitrators [][]byte
	BlockHeight uint32

	hash *common.Hash
}

func (i *InactiveArbitrators) Type() IllegalDataType {
	return InactiveArbitrator
}

func (i *InactiveArbitrators) GetBlockHeight() uint32 {
	return i.BlockHeight
}

func (i *InactiveArbitrators) Data(version byte) []byte {
	buf := new(bytes.Buffer)
	if err := i.Serialize(buf, version); err != nil {
		return []byte{0}
	}
	return buf.Bytes()
}

func (i *InactiveArbitrators) SerializeUnsigned(w io.Writer, version byte) error {
	return nil
}

func (i *InactiveArbitrators) Serialize(w io.Writer, version byte) error {
	return nil
}

func (i *InactiveArbitrators) DeserializeUnsigned(r io.Reader,
	version byte) (err error) {
	return err
}

func (i *InactiveArbitrators) Deserialize(r io.Reader,
	version byte) (err error) {
	return nil
}

func (i *InactiveArbitrators) Hash() common.Hash {
	if i.hash == nil {
		hash := rlpHash(i)
		i.hash = &hash
	}
	return *i.hash
}
