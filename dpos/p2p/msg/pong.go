// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dposp2p"
)

// Ensure Pong implement p2p.Message interface.
var _ dposp2p.Message = (*Pong)(nil)

type Pong struct {
	Ping
}

func NewPong(nonce uint64) *Pong {
	pong := new(Pong)
	pong.Nonce = nonce
	return pong
}

func (msg *Pong) CMD() string {
	return dposp2p.CmdPong
}