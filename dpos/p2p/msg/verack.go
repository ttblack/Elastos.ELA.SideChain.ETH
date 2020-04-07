// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dposp2p"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

// Ensure VerAck implement p2p.Message interface.
var _ dposp2p.Message = (*VerAck)(nil)

type VerAck struct {
	Signature [64]byte
}

func (msg *VerAck) CMD() string {
	return dposp2p.CmdVerAck
}

func (msg *VerAck) MaxLength() uint32 {
	return 64
}

func (msg *VerAck) Serialize(w io.Writer) error {
	_, err := w.Write(msg.Signature[:])
	return err
}

func (msg *VerAck) Deserialize(s *rlp.Stream) error {
	return nil
}

func NewVerAck(signature []byte) *VerAck {
	verAck := VerAck{}
	copy(verAck.Signature[:], signature)
	return &verAck
}
