// Copyright 2015 The Elastos.ELA.SideChain.ETH Authors
// This file is part of the Elastos.ELA.SideChain.ETH library.
//
// The Elastos.ELA.SideChain.ETH library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Elastos.ELA.SideChain.ETH library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Elastos.ELA.SideChain.ETH library. If not, see <http://www.gnu.org/licenses/>.

package msg_pool

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/chains/evm/voter"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

// Transactions is a Transaction slice type for basic sorting.
type RelayMsgs []*voter.Proposal

// Len returns the length of s.
func (s RelayMsgs) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s RelayMsgs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s RelayMsgs) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

func (s RelayMsgs) Less(i, j int) bool { return s[i].DepositNonce > s[j].DepositNonce }

type MsgPool struct {
	items map[uint64]*voter.Proposal // Hash map storing the transaction data
	lock sync.RWMutex
}

func NewMsgPool() *MsgPool {
	return &MsgPool{
		items: make(map[uint64]*voter.Proposal),
	}
}

// Get retrieves the current transactions associated with the given nonce.
func (m *MsgPool) Get(nonce uint64) *voter.Proposal {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.items[nonce]
}

func (m *MsgPool) Put(msg *voter.Proposal) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	nonce := msg.DepositNonce
	if m.items[nonce] != nil {
		return errors.New(fmt.Sprintf("error nonce %d", nonce))
	}
	m.items[nonce] = msg
	return nil
}

func (m *MsgPool) Flatten() []*voter.Proposal {
	m.lock.RLock()
	defer m.lock.RUnlock()

	txs := make([]*voter.Proposal, 0, len(m.items))
	for _, msg := range m.items {
		txs = append(txs, msg)
	}
	sort.Sort(RelayMsgs(txs))
	return txs
}

func (m *MsgPool) Remove(nonce uint64) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.items[nonce]
	if !ok {
		return false
	}
	delete(m.items, nonce)
	return true
}