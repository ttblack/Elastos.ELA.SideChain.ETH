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

const (
	// maxHostLength defines the maximum host name length.
	maxHostLength = 253
)

// Ensure Ping implement p2p.Message interface.
var _ dposp2p.Message = (*Addr)(nil)

// Addr represents a DPoS network address for connect to a peer.
type Addr struct {
	Host string
	Port uint16
}

func NewAddr(host string, port uint16) *Addr {
	return &Addr{Host: host, Port: port}
}

func (msg *Addr) CMD() string {
	return dposp2p.CmdAddr
}

func (msg *Addr) MaxLength() uint32 {
	return maxHostLength + 2
}

func (msg *Addr) Serialize(w io.Writer) error {
	return nil
}

func (msg *Addr) Deserialize(s *rlp.Stream) error {
	var err error

	return err
}
