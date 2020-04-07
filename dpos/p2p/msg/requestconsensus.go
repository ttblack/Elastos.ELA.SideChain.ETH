// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

type RequestConsensus struct {
	Height uint64
}

func (msg *RequestConsensus) CMD() string {
	return CmdRequestConsensus
}

func (msg *RequestConsensus) MaxLength() uint32 {
	return 4
}

func (msg *RequestConsensus) Serialize(w io.Writer) error {

	return nil
}

func (msg *RequestConsensus) Deserialize(s *rlp.Stream) error {
	var err error

	return err
}
