// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package state

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	com "github.com/elastos/Elastos.ELA/common"
)

type Arbitrators interface {
	Start()
	CheckDPOSIllegalTx(block *types.Block) error

	IsArbitrator(account common.Address) bool
	GetArbitrators() [][]byte
	GetCandidates() [][]byte
	GetNextArbitrators() [][]byte
	GetNextCandidates() [][]byte
	//GetNeedConnectArbiters() []peer.PID
	GetDutyIndexByHeight(height uint64) int
	GetDutyIndex() int

	GetCurrentRewardData() RewardData
	GetNextRewardData() RewardData
	GetArbitersRoundReward() map[com.Uint168]uint64
	GetFinalRoundChange() uint64
	IsInactiveMode() bool
	IsUnderstaffedMode() bool

	GetCRCArbiters() [][]byte
	GetCRCProducer(publicKey []byte) *Producer
	GetCRCArbitrators() map[string]*Producer
	IsCRCArbitrator(pk []byte) bool
	IsActiveProducer(pk []byte) bool
	IsDisabledProducer(pk []byte) bool

	GetOnDutyArbitrator() []byte
	GetNextOnDutyArbitrator(offset uint32) []byte

	GetOnDutyCrossChainArbitrator() []byte
	GetCrossChainArbiters() [][]byte
	GetCrossChainArbitersCount() int
	GetCrossChainArbitersMajorityCount() int

	GetArbitersCount() int
	GetCRCArbitersCount() int
	GetArbitersMajorityCount() int
	HasArbitersMajorityCount(num int) bool
	HasArbitersMinorityCount(num int) bool

	GetSnapshot(height uint64) []*KeyFrame
	DumpInfo(height uint64)
}

type IArbitratorsRecord interface {
	GetHeightsDesc() ([]uint64, error)
	GetCheckPoint(height uint64) (*CheckPoint, error)
	SaveArbitersState(point *CheckPoint) error
}
