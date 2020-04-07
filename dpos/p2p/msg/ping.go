// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"encoding/binary"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dposp2p"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

// Ensure Ping implement p2p.Message interface.
var _ dposp2p.Message = (*Ping)(nil)

type Ping struct {
	Nonce uint64
}

func NewPing(nonce uint64) *Ping {
	ping := new(Ping)
	ping.Nonce = nonce
	return ping
}

func (msg *Ping) CMD() string {
	return dposp2p.CmdPing
}

func (msg *Ping) MaxLength() uint32 {
	return 8
}

func (msg *Ping) Serialize(writer io.Writer) error {
	return binary.Write(writer, binary.LittleEndian, msg.Nonce)
}

func (msg *Ping) Deserialize(s *rlp.Stream) error {
	return nil
}
