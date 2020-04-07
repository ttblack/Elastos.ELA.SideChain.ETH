// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dposp2p"
	"io"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/dtime"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

// Ensure Version implement p2p.Message interface.
var _ dposp2p.Message = (*Version)(nil)

type Version struct {
	PID       [33]byte
	Target    [16]byte
	Nonce     [16]byte
	Port      uint16
	Timestamp time.Time
}

func (msg *Version) CMD() string {
	return dposp2p.CmdVersion
}

func (msg *Version) MaxLength() uint32 {
	return 75 // 33+16+16+2+8
}

func (msg *Version) Serialize(w io.Writer) error {
	return nil
}

func (msg *Version) Deserialize(s *rlp.Stream) error {
	return nil
}

func NewVersion(pid [33]byte, target, nonce [16]byte, port uint16) *Version {
	return &Version{PID: pid, Target: target, Nonce: nonce, Port: port,
		Timestamp:dtime.Now()}
}
