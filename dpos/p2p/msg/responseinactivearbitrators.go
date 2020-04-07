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

const ResponseInactiveArbitratorsLength = 32 + 33 + 64 + 8*2

type ResponseInactiveArbitrators struct {
	TxHash common.Hash
	Signer []byte
	Sign   []byte
}

func (i *ResponseInactiveArbitrators) CMD() string {
	return CmdResponseInactiveArbitrators
}

func (i *ResponseInactiveArbitrators) MaxLength() uint32 {
	return ResponseInactiveArbitratorsLength
}

func (i *ResponseInactiveArbitrators) Serialize(w io.Writer) error {
	return nil
}

func (i *ResponseInactiveArbitrators) SerializeUnsigned(w io.Writer) error {
	return nil
}

func (i *ResponseInactiveArbitrators) Deserialize(s *rlp.Stream) (err error) {
	return err
}

func (i *ResponseInactiveArbitrators) DeserializeUnsigned(r io.Reader) (err error) {
	return err
}
