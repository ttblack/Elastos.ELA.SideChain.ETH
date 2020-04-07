// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"io"
)

const DefaultResponseConsensusMessageDataSize = 8000000 * 10

type ResponseConsensus struct {
	Consensus ConsensusStatus
}

func (msg *ResponseConsensus) CMD() string {
	return CmdResponseConsensus
}

func (msg *ResponseConsensus) MaxLength() uint32 {
	return DefaultResponseConsensusMessageDataSize
}

func (msg *ResponseConsensus) Serialize(w io.Writer) error {
	return msg.Consensus.Serialize(w)
}

func (msg *ResponseConsensus) Deserialize(s *rlp.Stream) error {
	return msg.Consensus.Deserialize(s)
}
