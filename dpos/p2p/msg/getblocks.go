// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"
)

type GetBlocks struct {
	StartBlockHeight uint64
	EndBlockHeight   uint64
}

func (msg *GetBlocks) CMD() string {
	return CmdGetBlocks
}

func (msg *GetBlocks) MaxLength() uint32 {
	return 8
}

func (msg *GetBlocks) Serialize(w io.Writer) error {
	return nil
}

func (msg *GetBlocks) Deserialize(s *rlp.Stream) error {
	var err error

	return err
}
