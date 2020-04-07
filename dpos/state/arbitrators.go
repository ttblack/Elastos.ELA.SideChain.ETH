// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package state

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/params"
	"github.com/elastos/Elastos.ELA/core/contract"
	"github.com/elastos/Elastos.ELA/core/types/payload"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/events"
	dtypes "github.com/elastos/Elastos.ELA.SideChain.ETH/dpos/types"
	com "github.com/elastos/Elastos.ELA/common"
	epeer "github.com/elastos/Elastos.ELA/dpos/p2p/peer"
)

type ChangeType byte

const (
	// MajoritySignRatioNumerator defines the ratio numerator to achieve
	// majority signatures.
	MajoritySignRatioNumerator = float64(2)

	// MajoritySignRatioDenominator defines the ratio denominator to achieve
	// majority signatures.
	MajoritySignRatioDenominator = float64(3)

	// MaxNormalInactiveChangesCount defines the max count arbitrators can
	// change when more than 1/3 arbiters don't sign cause to confirm fail
	MaxNormalInactiveChangesCount = 3

	// MaxSnapshotLength defines the max length the snapshot map should take
	MaxSnapshotLength = 20

	none         = ChangeType(0x00)
	updateNext   = ChangeType(0x01)
	normalChange = ChangeType(0x02)
)

var (
	ErrInsufficientProducer = errors.New("producers count less than min arbitrators count")
)

type arbitrators struct {
	*State
	*degradation
	*KeyFrame
	getBlockByHeight func(uint64) (*types.Block, error)
	bestHeight func() uint64

	mtx               sync.Mutex
	started           bool
	dutyIndex         int
	currentCandidates [][]byte

	CurrentReward RewardData
	NextReward    RewardData

	nextArbitrators             [][]byte
	nextCandidates              [][]byte
	crcArbiters                 [][]byte
	crcArbitratorsProgramHashes map[com.Uint168]interface{}
	crcArbitratorsNodePublicKey map[string]*Producer
	accumulativeReward          uint64
	finalRoundChange            uint64
	clearingHeight              uint64
	arbitersRoundReward         map[com.Uint168]uint64
	illegalBlocksPayloadHashes  map[common.Hash]interface{}

	snapshots            map[uint64][]*CheckPoint
	snapshotKeysDesc     []uint64
	lastCheckPointHeight uint64

	forceChanged bool
}

func (a *arbitrators) Start() {
	a.mtx.Lock()
	a.started = true
	a.mtx.Unlock()
}

func (a *arbitrators) RegisterFunction(bestHeight func() uint64,
	getBlockByHeight func(uint64) (*types.Block, error)) {
	a.bestHeight = bestHeight
	a.getBlockByHeight = getBlockByHeight
}

func (a *arbitrators) RecoverFromCheckPoints(point *CheckPoint) {
	a.mtx.Lock()
	a.recoverFromCheckPoints(point)
	a.mtx.Unlock()
}

func (a *arbitrators) recoverFromCheckPoints(point *CheckPoint) {
	a.dutyIndex = point.DutyIndex
	a.CurrentArbitrators = point.CurrentArbitrators
	a.currentCandidates = point.CurrentCandidates
	a.nextArbitrators = point.NextArbitrators
	a.nextCandidates = point.NextCandidates
	a.CurrentReward = point.CurrentReward
	a.NextReward = point.NextReward
	a.StateKeyFrame = &point.StateKeyFrame
	a.accumulativeReward = point.accumulativeReward
	a.finalRoundChange = point.finalRoundChange
	a.clearingHeight = point.clearingHeight
	a.arbitersRoundReward = point.arbitersRoundReward
	a.illegalBlocksPayloadHashes = point.illegalBlocksPayloadHashes
}

func (a *arbitrators) ProcessBlock(block *types.Block, confirm *payload.Confirm) {
	a.State.ProcessBlock(block, confirm)
	a.IncreaseChainHeight(block)
}

func (a *arbitrators) CheckDPOSIllegalTx(block *types.Block) error {

	a.mtx.Lock()
	hashes := a.illegalBlocksPayloadHashes
	a.mtx.Unlock()

	if hashes == nil || len(hashes) == 0 {
		return nil
	}

	foundMap := make(map[common.Hash]bool)
	for k := range hashes {
		foundMap[k] = false
	}

	//for _, tx := range block.Transactions {
	//	if tx.IsIllegalBlockTx() {
	//		foundMap[tx.Payload.(*payload.DPOSIllegalBlocks).Hash()] = true
	//	}
	//}

	for _, found := range foundMap {
		if !found {
			return errors.New("expect an illegal blocks transaction in this block")
		}
	}
	return nil
}

//func (a *arbitrators) ProcessSpecialTxPayload(p types.Payload,
//	height uint32) error {
//	switch obj := p.(type) {
//	case *payload.DPOSIllegalBlocks:
//		a.mtx.Lock()
//		a.illegalBlocksPayloadHashes[obj.Hash()] = nil
//		a.mtx.Unlock()
//	case *payload.InactiveArbitrators:
//		if !a.AddInactivePayload(obj) {
//			log.Debug("[ProcessSpecialTxPayload] duplicated payload")
//			return nil
//		}
//	default:
//		return errors.New("[ProcessSpecialTxPayload] invalid payload type")
//	}
//
//	a.State.ProcessSpecialTxPayload(p, height)
//	return a.ForceChange(height)
//}

func (a *arbitrators) RollbackTo(height uint64) error {
	if height > a.history.Height() {
		return fmt.Errorf("can't rollback to height: %d", height)
	}

	if err := a.DecreaseChainHeight(height); err != nil {
		return err
	}

	return a.State.RollbackTo(height)
}

func (a *arbitrators) GetDutyIndexByHeight(height uint64) (index int) {
	a.mtx.Lock()
	//if height >= a.chainParams.CRCOnlyDPOSHeight-1 {
	//	index = a.dutyIndex % len(a.CurrentArbitrators)
	//} else {
	//	index = int(height) % len(a.CurrentArbitrators)
	//}
	a.mtx.Unlock()
	return index
}

func (a *arbitrators) GetDutyIndex() int {
	a.mtx.Lock()
	index := a.dutyIndex
	a.mtx.Unlock()

	return index
}

func (a *arbitrators) GetArbitersRoundReward() map[com.Uint168]uint64 {
	a.mtx.Lock()
	result := a.arbitersRoundReward
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetFinalRoundChange() uint64 {
	a.mtx.Lock()
	result := a.finalRoundChange
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) ForceChange(height uint64) error {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	block, err := a.getBlockByHeight(height)
	if err != nil {
		block, err = a.getBlockByHeight(a.bestHeight())
		if err != nil {
			return err
		}
	}
	a.snapshot(height)

	if err := a.clearingDPOSReward(block, false); err != nil {
		return err
	}

	if err := a.updateNextArbitrators(height + 1); err != nil {
		return err
	}

	if err := a.changeCurrentArbitrators(); err != nil {
		return err
	}

	if a.started {
		go events.Notify(events.ETDirectPeersChanged,
			a.getNeedConnectArbiters())
	}

	a.forceChanged = true

	a.dumpInfo(height)
	return nil
}

func (a *arbitrators) tryHandleError(height uint64, err error) error {
	if err == ErrInsufficientProducer {
		log.Warn("found error: ", err, ", degrade to CRC only state")
		a.TrySetUnderstaffed(height)
		return nil
	} else {
		return err
	}
}

func (a *arbitrators) normalChange(height uint64) error {
	if err := a.changeCurrentArbitrators(); err != nil {
		log.Warn("[NormalChange] change current arbiters error: ", err)
		return err
	}

	if err := a.updateNextArbitrators(height + 1); err != nil {
		log.Warn("[NormalChange] update next arbiters error: ", err)
		return err
	}

	return nil
}

func (a *arbitrators) IncreaseChainHeight(block *types.Block) {
	var notify = true

	a.mtx.Lock()

	changeType, versionHeight := a.getChangeType(block.NumberU64() + 1)
	switch changeType {
	case updateNext:
		if err := a.updateNextArbitrators(versionHeight); err != nil {
			panic(fmt.Sprintf("[IncreaseChainHeight] update next arbiters at height: %d, error: %s", block.NumberU64(), err))
		}
	case normalChange:
		if err := a.clearingDPOSReward(block, true); err != nil {
			panic(fmt.Sprintf("normal change fail when clear DPOS reward: "+
				" transaction, height: %d, error: %s", block.NumberU64(), err))
		}
		if err := a.normalChange(block.NumberU64()); err != nil {
			panic(fmt.Sprintf("normal change fail at height: %d, error: %s",
				block.NumberU64(), err))
		}
	case none:
		a.accumulateReward(block)
		a.dutyIndex++
		notify = false
	}
	a.illegalBlocksPayloadHashes = make(map[common.Hash]interface{})

	if block.NumberU64() > a.bestHeight()-MaxSnapshotLength {
		a.snapshot(block.NumberU64())
	}

	a.mtx.Unlock()

	if a.started && notify {
		go events.Notify(events.ETDirectPeersChanged, a.GetNeedConnectArbiters())
	}
}

func (a *arbitrators) accumulateReward(block *types.Block) {
	//if block.Height < a.State.chainParams.PublicDPOSHeight {
	//	return
	//}
	//
	//if block.Height < a.chainParams.CRVotingStartHeight || !a.forceChanged {
	//	dposReward := a.getBlockDPOSReward(block)
	//	a.accumulativeReward += dposReward
	//}
	//
	//a.arbitersRoundReward = nil
	//a.finalRoundChange = 0
	//a.forceChanged = false
}

func (a *arbitrators) clearingDPOSReward(block *types.Block,
	smoothClearing bool) error {
	//if block.Height < a.chainParams.PublicDPOSHeight ||
	//	block.Height == a.clearingHeight {
	//	return nil
	//}

	dposReward := a.getBlockDPOSReward(block)
	if smoothClearing {
		a.accumulativeReward += dposReward
		dposReward = 0
	}

	if err := a.distributeDPOSReward(a.accumulativeReward); err != nil {
		return err
	}
	a.accumulativeReward = dposReward
	a.clearingHeight = block.NumberU64()

	return nil
}

func (a *arbitrators) distributeDPOSReward(reward uint64) (err error) {
	a.arbitersRoundReward = map[com.Uint168]uint64{}

	//a.arbitersRoundReward[a.chainParams.CRCAddress] = 0
	//realDPOSReward, err := a.distributeWithNormalArbitrators(reward)
	//
	//if err != nil {
	//	return err
	//}
	//
	//change := reward - realDPOSReward
	//if change < 0 {
	//	return errors.New("real dpos reward more than reward limit")
	//}
	//
	//a.finalRoundChange = change
	return nil
}

func (a *arbitrators) distributeWithNormalArbitrators(
	reward uint64) (uint64, error) {
	ownerHashes := a.CurrentReward.OwnerProgramHashes
	if len(ownerHashes) == 0 {
		return 0, errors.New("not found arbiters when distributeWithNormalArbitrators")
	}

	//totalBlockConfirmReward := float64(reward) * 0.25
	//totalTopProducersReward := float64(reward) - totalBlockConfirmReward
	//individualBlockConfirmReward := common.Fixed64(
	//	math.Floor(totalBlockConfirmReward / float64(len(ownerHashes))))
	//totalVotesInRound := a.CurrentReward.TotalVotesInRound
	//if len(a.chainParams.CRCArbiters) == len(a.CurrentArbitrators) {
	//	a.arbitersRoundReward[a.chainParams.CRCAddress] = reward
	//	return reward, nil
	//}
	//rewardPerVote := totalTopProducersReward / float64(totalVotesInRound)
	//
	//realDPOSReward := common.Fixed64(0)
	//for _, ownerHash := range ownerHashes {
	//	votes := a.CurrentReward.OwnerVotesInRound[*ownerHash]
	//	individualProducerReward := common.Fixed64(math.Floor(float64(
	//		votes) * rewardPerVote))
	//	r := individualBlockConfirmReward + individualProducerReward
	//	if _, ok := a.crcArbitratorsProgramHashes[*ownerHash]; ok {
	//		r = individualBlockConfirmReward
	//		a.arbitersRoundReward[a.chainParams.CRCAddress] += r
	//	} else {
	//		a.arbitersRoundReward[*ownerHash] = r
	//	}
	//
	//	realDPOSReward += r
	//}
	//candidateOwnerHashes := a.CurrentReward.CandidateOwnerProgramHashes
	//for _, ownerHash := range candidateOwnerHashes {
	//	votes := a.CurrentReward.OwnerVotesInRound[*ownerHash]
	//	individualProducerReward := common.Fixed64(math.Floor(float64(
	//		votes) * rewardPerVote))
	//	a.arbitersRoundReward[*ownerHash] = individualProducerReward
	//
	//	realDPOSReward += individualProducerReward
	//}
	//return realDPOSReward, nil
	return 0, nil
}

func (a *arbitrators) DecreaseChainHeight(height uint64) error {
	a.mtx.Lock()
	a.degradation.RollbackTo(height)

	checkpoints := a.getSnapshot(height)
	if checkpoints != nil {
		for _, v := range checkpoints {
			a.recoverFromCheckPoints(v)
			// since we don't support force change at now, length of checkpoints
			// should be 1, so we break the loop here
			break
		}
	}
	a.mtx.Unlock()

	return nil
}

func (a *arbitrators) GetNeedConnectArbiters() []epeer.PID {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	return a.getNeedConnectArbiters()
}

func (a *arbitrators) getNeedConnectArbiters() []epeer.PID {
	//height := a.history.Height() + 1
	//if height < a.chainParams.CRCOnlyDPOSHeight-a.chainParams.PreConnectOffset {
	//	return nil
	//}
	//
	//pids := make(map[string]peer.PID)
	//for k, p := range a.crcArbitratorsNodePublicKey {
	//	var pid peer.PID
	//	copy(pid[:], p.NodePublicKey())
	//	pids[k] = pid
	//}
	//
	//for _, v := range a.CurrentArbitrators {
	//	key := common.BytesToHexString(v)
	//	var pid peer.PID
	//	copy(pid[:], v)
	//	pids[key] = pid
	//}
	//
	//for _, v := range a.nextArbitrators {
	//	key := common.BytesToHexString(v)
	//	var pid peer.PID
	//	copy(pid[:], v)
	//	pids[key] = pid
	//}
	//
	//peers := make([]peer.PID, 0, len(pids))
	//for _, pid := range pids {
	//	peers = append(peers, pid)
	//}
	//
	//return peers
	return nil
}

func (a *arbitrators) IsArbitrator(account common.Address) bool {
	arbitrators := a.GetArbitrators()
	for _, v := range arbitrators {
		if bytes.Equal(account.Bytes(), v) {
			return true
		}
	}
	return false
}

func (a *arbitrators) GetArbitrators() [][]byte {
	a.mtx.Lock()
	result := a.CurrentArbitrators
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetCandidates() [][]byte {
	a.mtx.Lock()
	result := a.currentCandidates
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetNextArbitrators() [][]byte {
	a.mtx.Lock()
	result := a.nextArbitrators
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetNextCandidates() [][]byte {
	a.mtx.Lock()
	result := a.nextCandidates
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetCRCArbiters() [][]byte {
	a.mtx.Lock()
	result := a.crcArbiters
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetCurrentRewardData() RewardData {
	a.mtx.Lock()
	result := a.CurrentReward
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) GetNextRewardData() RewardData {
	a.mtx.Lock()
	result := a.NextReward
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) IsCRCArbitrator(pk []byte) bool {
	// there is no need to lock because crc related variable is read only and
	// initialized at the very first
	_, ok := a.crcArbitratorsNodePublicKey[hex.EncodeToString(pk)]
	return ok
}

func (a *arbitrators) IsActiveProducer(pk []byte) bool {
	return a.State.IsActiveProducer(pk)
}

func (a *arbitrators) IsDisabledProducer(pk []byte) bool {
	return a.State.IsInactiveProducer(pk) || a.State.IsIllegalProducer(pk) || a.State.IsCanceledProducer(pk)
}

func (a *arbitrators) GetCRCProducer(publicKey []byte) *Producer {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	key := hex.EncodeToString(publicKey)
	if producer, ok := a.crcArbitratorsNodePublicKey[key]; ok {
		return producer
	}
	return nil
}

func (a *arbitrators) GetCRCArbitrators() map[string]*Producer {
	return a.crcArbitratorsNodePublicKey
}

func (a *arbitrators) GetOnDutyArbitrator() []byte {
	return a.GetNextOnDutyArbitratorV(a.bestHeight()+1, 0)
}

func (a *arbitrators) GetNextOnDutyArbitrator(offset uint32) []byte {
	return a.GetNextOnDutyArbitratorV(a.bestHeight()+1, offset)
}

func (a *arbitrators) GetOnDutyCrossChainArbitrator() []byte {
	var arbiter []byte
	//height := a.bestHeight()
	//if height < a.chainParams.CRCOnlyDPOSHeight-1 {
	//	arbiter = a.GetOnDutyArbitrator()
	//} else {
	//	crcArbiters := a.GetCRCArbiters()
	//	sort.Slice(crcArbiters, func(i, j int) bool {
	//		return bytes.Compare(crcArbiters[i], crcArbiters[j]) < 0
	//	})
	//	ondutyIndex := int(height-a.chainParams.CRCOnlyDPOSHeight+1) % len(crcArbiters)
	//	arbiter = crcArbiters[ondutyIndex]
	//}

	return arbiter
}

func (a *arbitrators) GetCrossChainArbiters() [][]byte {
	//if a.bestHeight() < a.chainParams.CRCOnlyDPOSHeight-1 {
	//	return a.GetArbitrators()
	//}
	return a.GetCRCArbiters()
}

func (a *arbitrators) GetCrossChainArbitersCount() int {
	//if a.bestHeight() < a.chainParams.CRCOnlyDPOSHeight-1 {
	//	return len(a.chainParams.OriginArbiters)
	//}
	//
	//return len(a.chainParams.CRCArbiters)
	return 0
}

func (a *arbitrators) GetCrossChainArbitersMajorityCount() int {
	minSignCount := int(float64(a.GetCrossChainArbitersCount()) *
		MajoritySignRatioNumerator / MajoritySignRatioDenominator)
	return minSignCount
}

func (a *arbitrators) GetNextOnDutyArbitratorV(height uint64, offset uint32) []byte {
	// main version is >= H1
	//if height >= a.State.chainParams.CRCOnlyDPOSHeight {
	//	arbitrators := a.CurrentArbitrators
	//	if len(arbitrators) == 0 {
	//		return nil
	//	}
	//	index := (a.dutyIndex + int(offset)) % len(arbitrators)
	//	arbiter := arbitrators[index]
	//
	//	return arbiter
	//}

	// old version
	return a.getNextOnDutyArbitratorV0(height, offset)
}

func (a *arbitrators) GetArbitersCount() int {
	a.mtx.Lock()
	result := len(a.CurrentArbitrators)
	a.mtx.Unlock()
	return result
}

func (a *arbitrators) GetCRCArbitersCount() int {
	a.mtx.Lock()
	result := len(a.crcArbiters)
	a.mtx.Unlock()
	return result
}

func (a *arbitrators) GetArbitersMajorityCount() int {
	a.mtx.Lock()
	minSignCount := int(float64(len(a.CurrentArbitrators)) *
		MajoritySignRatioNumerator / MajoritySignRatioDenominator)
	a.mtx.Unlock()
	return minSignCount
}

func (a *arbitrators) HasArbitersMajorityCount(num int) bool {
	return num > a.GetArbitersMajorityCount()
}

func (a *arbitrators) HasArbitersMinorityCount(num int) bool {
	a.mtx.Lock()
	count := len(a.CurrentArbitrators)
	a.mtx.Unlock()
	return num >= count-a.GetArbitersMajorityCount()
}

func (a *arbitrators) getChangeType(height uint64) (ChangeType, uint64) {

	// special change points:
	//		H1 - PreConnectOffset -> 	[updateNext, H1]: update next arbiters and let CRC arbiters prepare to connect
	//		H1 -> 						[normalChange, H1]: should change to new election (that only have CRC arbiters)
	//		H2 - PreConnectOffset -> 	[updateNext, H2]: update next arbiters and let normal arbiters prepare to connect
	//		H2 -> 						[normalChange, H2]: should change to new election (arbiters will have both CRC and normal arbiters)
	//if height == a.State.chainParams.CRCOnlyDPOSHeight-
	//	a.State.chainParams.PreConnectOffset {
	//	return updateNext, a.State.chainParams.CRCOnlyDPOSHeight
	//} else if height == a.State.chainParams.CRCOnlyDPOSHeight {
	//	return normalChange, a.State.chainParams.CRCOnlyDPOSHeight
	//} else if height == a.State.chainParams.PublicDPOSHeight-
	//	a.State.chainParams.PreConnectOffset {
	//	return updateNext, a.State.chainParams.PublicDPOSHeight
	//} else if height == a.State.chainParams.PublicDPOSHeight {
	//	return normalChange, a.State.chainParams.PublicDPOSHeight
	//}
	//
	//// main version >= H2
	//if height > a.State.chainParams.PublicDPOSHeight &&
	//	a.dutyIndex == len(a.CurrentArbitrators)-1 {
	//	return normalChange, height
	//}

	return none, height
}

func (a *arbitrators) changeCurrentArbitrators() error {
	a.CurrentArbitrators = a.nextArbitrators
	a.currentCandidates = a.nextCandidates
	a.CurrentReward = a.NextReward

	sort.Slice(a.CurrentArbitrators, func(i, j int) bool {
		return bytes.Compare(a.CurrentArbitrators[i], a.CurrentArbitrators[j]) < 0
	})

	a.dutyIndex = 0
	return nil
}

func (a *arbitrators) updateNextArbitrators(height uint64) error {
	_, recover := a.InactiveModeSwitch(height, a.IsAbleToRecoverFromInactiveMode)
	if recover {
		a.LeaveEmergency()
	} else {
		a.TryLeaveUnderStaffed(a.IsAbleToRecoverFromUnderstaffedState)
	}

	a.nextArbitrators = make([][]byte, 0)
	for _, v := range a.crcArbitratorsNodePublicKey {
		a.nextArbitrators = append(a.nextArbitrators, v.info.NodePublicKey)
	}

	if !a.IsInactiveMode() && !a.IsUnderstaffedMode() {
		count := 3//a.chainParams.GeneralArbiters
		votedProducers := a.State.GetVotedProducers()
		sort.Slice(votedProducers, func(i, j int) bool {
			if votedProducers[i].votes == votedProducers[j].votes {
				return bytes.Compare(votedProducers[i].info.NodePublicKey,
					votedProducers[j].NodePublicKey()) < 0
			}
			return votedProducers[i].Votes() > votedProducers[j].Votes()
		})

		producers, err := a.GetNormalArbitratorsDesc(height, count,
			votedProducers)
		if err != nil {
			if err := a.tryHandleError(height, err); err != nil {
				return err
			}
			a.nextCandidates = make([][]byte, 0)
		} else {
			a.nextArbitrators = append(a.nextArbitrators, producers...)

			candidates, err := a.GetCandidatesDesc(height, count,
				votedProducers)
			if err != nil {
				return err
			}
			a.nextCandidates = candidates
		}
	} else {
		a.nextCandidates = make([][]byte, 0)
	}

	if err := a.snapshotVotesStates(); err != nil {
		return err
	}
	if err := a.updateNextOwnerProgramHashes(); err != nil {
		return err
	}

	return nil
}

func (a *arbitrators) GetCandidatesDesc(height uint64, startIndex int,
	producers []*Producer) ([][]byte, error) {
	// main version >= H2
	//if height >= a.State.chainParams.PublicDPOSHeight {
	//	if len(producers) < startIndex {
	//		return make([][]byte, 0), nil
	//	}
	//
	//	result := make([][]byte, 0)
	//	for i := startIndex; i < len(producers) && i < startIndex+a.
	//		chainParams.CandidateArbiters; i++ {
	//		result = append(result, producers[i].NodePublicKey())
	//	}
	//	return result, nil
	//}

	// old version [0, H2)
	return nil, nil
}

func (a *arbitrators) GetNormalArbitratorsDesc(height uint64,
	arbitratorsCount int, producers []*Producer) ([][]byte, error) {
	// main version >= H2
	//if height >= a.State.chainParams.PublicDPOSHeight {
	//	if len(producers) < arbitratorsCount {
	//		return nil, ErrInsufficientProducer
	//	}
	//
	//	result := make([][]byte, 0)
	//	for i := 0; i < arbitratorsCount && i < len(producers); i++ {
	//		result = append(result, producers[i].NodePublicKey())
	//	}
	//	return result, nil
	//}
	//
	//// version [H1, H2)
	//if height >= a.State.chainParams.CRCOnlyDPOSHeight {
	//	return a.getNormalArbitratorsDescV1()
	//}

	// version [0, H1)
	return a.getNormalArbitratorsDescV0()
}

func (a *arbitrators) snapshotVotesStates() error {
	a.NextReward.OwnerVotesInRound = make(map[com.Uint168]uint64, 0)
	a.NextReward.TotalVotesInRound = 0
	for _, nodePublicKey := range a.nextArbitrators {
		if !a.IsCRCArbitrator(nodePublicKey) {
			producer := a.GetProducer(nodePublicKey)
			if producer == nil {
				return errors.New("get producer by node public key failed")
			}
			programHash, err := contract.PublicKeyToStandardProgramHash(
				producer.OwnerPublicKey())
			if err != nil {
				return err
			}
			a.NextReward.OwnerVotesInRound[*programHash] = producer.Votes()
			a.NextReward.TotalVotesInRound += producer.Votes()
		}
	}

	for _, nodePublicKey := range a.nextCandidates {
		if a.IsCRCArbitrator(nodePublicKey) {
			continue
		}
		producer := a.GetProducer(nodePublicKey)
		if producer == nil {
			return errors.New("get producer by node public key failed")
		}
		programHash, err := contract.PublicKeyToStandardProgramHash(producer.OwnerPublicKey())
		if err != nil {
			return err
		}
		a.NextReward.OwnerVotesInRound[*programHash] = producer.Votes()
		a.NextReward.TotalVotesInRound += producer.Votes()
	}

	return nil
}

func (a *arbitrators) updateNextOwnerProgramHashes() error {
	a.NextReward.OwnerProgramHashes = make([]*com.Uint168, 0)
	for _, nodePublicKey := range a.nextArbitrators {
		if a.IsCRCArbitrator(nodePublicKey) {
			ownerPublicKey := nodePublicKey // crc node public key is its owner public key for now
			programHash, err := contract.PublicKeyToStandardProgramHash(ownerPublicKey)
			if err != nil {
				return err
			}
			a.NextReward.OwnerProgramHashes = append(
				a.NextReward.OwnerProgramHashes, programHash)
		} else {
			producer := a.GetProducer(nodePublicKey)
			if producer == nil {
				return errors.New("get producer by node public key failed")
			}
			ownerPublicKey := producer.OwnerPublicKey()
			programHash, err := contract.PublicKeyToStandardProgramHash(ownerPublicKey)
			if err != nil {
				return err
			}
			a.NextReward.OwnerProgramHashes = append(
				a.NextReward.OwnerProgramHashes, programHash)
		}
	}

	a.NextReward.CandidateOwnerProgramHashes = make([]*com.Uint168, 0)
	for _, nodePublicKey := range a.nextCandidates {
		if a.IsCRCArbitrator(nodePublicKey) {
			continue
		}
		producer := a.GetProducer(nodePublicKey)
		if producer == nil {
			return errors.New("get producer by node public key failed")
		}
		programHash, err := contract.PublicKeyToStandardProgramHash(producer.OwnerPublicKey())
		if err != nil {
			return err
		}
		a.NextReward.CandidateOwnerProgramHashes = append(
			a.NextReward.CandidateOwnerProgramHashes, programHash)
	}

	return nil
}

func (a *arbitrators) DumpInfo(height uint64) {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.dumpInfo(height)
}

func (a *arbitrators) dumpInfo(height uint64) {
	var printer func(string, ...interface{})
	changeType, _ := a.getChangeType(height + 1)
	switch changeType {
	case updateNext:
		fallthrough
	case normalChange:
		printer = log.Info
	case none:
		printer = log.Debug
	}

	var crInfo string
	crParams := make([]interface{}, 0)
	if len(a.CurrentArbitrators) != 0 {
		crInfo, crParams = getArbitersInfoWithOnduty("CURRENT ARBITERS",
			a.CurrentArbitrators, a.dutyIndex, a.GetOnDutyArbitrator())
	} else {
		crInfo, crParams = getArbitersInfoWithoutOnduty("CURRENT ARBITERS", a.CurrentArbitrators)
	}
	nrInfo, nrParams := getArbitersInfoWithoutOnduty("NEXT ARBITERS", a.nextArbitrators)
	ccInfo, ccParams := getArbitersInfoWithoutOnduty("CURRENT CANDIDATES", a.currentCandidates)
	ncInfo, ncParams := getArbitersInfoWithoutOnduty("NEXT CANDIDATES", a.nextCandidates)
	printer(crInfo+nrInfo+ccInfo+ncInfo, append(append(append(crParams, nrParams...), ccParams...), ncParams...)...)
}

func (a *arbitrators) getBlockDPOSReward(block *types.Block) uint64 {
	totalTxFx := big.NewInt(0)
	for _, tx := range block.Transactions() {
		//totalTxFx += tx.Gas() * tx.GasPrice()
		fee := tx.Gas() * tx.GasPrice().Uint64()
		totalTxFx.Add(totalTxFx, big.NewInt(int64(fee)))
	}

	//return (math.Ceil(float64(totalTxFx+
	//	//	a.chainParams.RewardPerBlock) * 0.35))
	return totalTxFx.Uint64()
}

func (a *arbitrators) newCheckPoint(height uint64) *CheckPoint {
	point := &CheckPoint{
		Height:                     height,
		DutyIndex:                  a.dutyIndex,
		CurrentCandidates:          make([][]byte, 0),
		NextArbitrators:            make([][]byte, 0),
		NextCandidates:             make([][]byte, 0),
		CurrentReward:              *NewRewardData(),
		NextReward:                 *NewRewardData(),
		accumulativeReward:         a.accumulativeReward,
		finalRoundChange:           a.finalRoundChange,
		clearingHeight:             a.clearingHeight,
		arbitersRoundReward:        make(map[com.Uint168]uint64),
		illegalBlocksPayloadHashes: make(map[common.Hash]interface{}),
		KeyFrame: KeyFrame{
			CurrentArbitrators: a.CurrentArbitrators,
		},
		StateKeyFrame: *a.State.snapshot(),
	}
	point.CurrentArbitrators = copyByteList(a.CurrentArbitrators)
	point.CurrentCandidates = copyByteList(a.currentCandidates)
	point.NextArbitrators = copyByteList(a.nextArbitrators)
	point.NextCandidates = copyByteList(a.nextCandidates)
	point.CurrentReward = *copyReward(&a.CurrentReward)
	point.NextReward = *copyReward(&a.NextReward)
	for k, v := range a.arbitersRoundReward {
		point.arbitersRoundReward[k] = v
	}
	for k := range a.illegalBlocksPayloadHashes {
		point.illegalBlocksPayloadHashes[k] = nil
	}

	return point
}

func (a *arbitrators) snapshot(height uint64) {
	var frames []*CheckPoint
	if v, ok := a.snapshots[height]; ok {
		frames = v
	} else {
		// remove the oldest keys if snapshot capacity is over
		if len(a.snapshotKeysDesc) >= MaxSnapshotLength {
			for i := MaxSnapshotLength - 1; i < len(a.snapshotKeysDesc); i++ {
				delete(a.snapshots, a.snapshotKeysDesc[i])
			}
			a.snapshotKeysDesc = a.snapshotKeysDesc[0 : MaxSnapshotLength-1]
		}

		a.snapshotKeysDesc = append(a.snapshotKeysDesc, height)
		sort.Slice(a.snapshotKeysDesc, func(i, j int) bool {
			return a.snapshotKeysDesc[i] > a.snapshotKeysDesc[j]
		})
	}
	checkpoint := a.newCheckPoint(height)
	frames = append(frames, checkpoint)
	a.snapshots[height] = frames
}

func (a *arbitrators) GetSnapshot(height uint64) (result []*KeyFrame) {
	a.mtx.Lock()
	if height > a.bestHeight() {
		// if height is larger than first snapshot then return current key frame
		result = append(result, a.KeyFrame)
	} else if height >= a.snapshotKeysDesc[len(a.snapshotKeysDesc)-1] {
		checkpoints := a.getSnapshot(height)
		result = make([]*KeyFrame, 0, len(checkpoints))
		for _, v := range checkpoints {
			result = append(result, &v.KeyFrame)
		}
	}
	a.mtx.Unlock()

	return result
}

func (a *arbitrators) getSnapshot(height uint64) []*CheckPoint {
	result := make([]*CheckPoint, 0)
	if height >= a.snapshotKeysDesc[len(a.snapshotKeysDesc)-1] {
		// if height is in range of snapshotKeysDesc, get the key with the same
		// election as height
		key := a.snapshotKeysDesc[0]
		for i := 1; i < len(a.snapshotKeysDesc); i++ {
			if height >= a.snapshotKeysDesc[i] &&
				height < a.snapshotKeysDesc[i-1] {
				key = a.snapshotKeysDesc[i]
			}
		}

		return a.snapshots[key]
	}
	return result
}

func getArbitersInfoWithOnduty(title string, arbiters [][]byte,
	dutyIndex int, ondutyArbiter []byte) (string, []interface{}) {
	info := "\n" + title + "\nDUTYINDEX: %d\n%5s %66s %6s \n----- " +
		strings.Repeat("-", 66) + " ------\n"
	params := make([]interface{}, 0)
	params = append(params, (dutyIndex+1)%len(arbiters))
	params = append(params, "INDEX", "PUBLICKEY", "ONDUTY")
	for i, arbiter := range arbiters {
		info += "%-5d %-66s %6t\n"
		publicKey := common.Bytes2Hex(arbiter)
		params = append(params, i+1, publicKey, bytes.Equal(arbiter, ondutyArbiter))
	}
	info += "----- " + strings.Repeat("-", 66) + " ------"
	return info, params
}

func getArbitersInfoWithoutOnduty(title string, arbiters [][]byte) (string, []interface{}) {

	info := "\n" + title + "\n%5s %66s\n----- " + strings.Repeat("-", 66) + "\n"
	params := make([]interface{}, 0)
	params = append(params, "INDEX", "PUBLICKEY")
	for i, arbiter := range arbiters {
		info += "%-5d %-66s\n"
		publicKey := common.Bytes2Hex(arbiter)
		params = append(params, i+1, publicKey)
	}
	info += "----- " + strings.Repeat("-", 66)
	return info, params
}

func (a *arbitrators) initArbitrators(chainParams *params.DposConfig) error {
	originArbiters := make([][]byte, len(chainParams.OriginArbiters))
	originArbitersProgramHashes := make([]*com.Uint168, len(chainParams.OriginArbiters))
	for i, arbiter := range chainParams.OriginArbiters {
		a := common.Hex2Bytes(arbiter)
		originArbiters[i] = a
		hash, err := contract.PublicKeyToStandardProgramHash(a)
		if err != nil {
			return err
		}
		originArbitersProgramHashes[i] = hash
	}

	crcNodeMap := make(map[string]*Producer)
	crcArbitratorsProgramHashes := make(map[com.Uint168]interface{})
	crcArbiters := make([][]byte, 0, len(chainParams.CRCArbiters))
	for _, pk := range chainParams.CRCArbiters {
		pubKey, err := hex.DecodeString(pk)
		if err != nil {
			return err
		}
		hash, err := contract.PublicKeyToStandardProgramHash(pubKey)
		if err != nil {
			return err
		}
		crcArbiters = append(crcArbiters, pubKey)
		crcArbitratorsProgramHashes[*hash] = nil
		crcNodeMap[pk] = &Producer{ // here need crc NODE public key
			info: dtypes.ProducerInfo{
				OwnerPublicKey: pubKey,
				NodePublicKey:  pubKey,
			},
			activateRequestHeight: math.MaxUint32,
		}
	}

	a.nextArbitrators = originArbiters
	a.crcArbiters = crcArbiters
	a.crcArbitratorsNodePublicKey = crcNodeMap
	a.crcArbitratorsProgramHashes = crcArbitratorsProgramHashes
	a.KeyFrame = &KeyFrame{CurrentArbitrators: originArbiters}
	a.CurrentReward = RewardData{
		OwnerProgramHashes:          originArbitersProgramHashes,
		CandidateOwnerProgramHashes: make([]*com.Uint168, 0),
		OwnerVotesInRound:           make(map[com.Uint168]uint64),
		TotalVotesInRound:           0,
	}
	a.NextReward = RewardData{
		OwnerProgramHashes:          originArbitersProgramHashes,
		CandidateOwnerProgramHashes: make([]*com.Uint168, 0),
		OwnerVotesInRound:           make(map[com.Uint168]uint64),
		TotalVotesInRound:           0,
	}
	return nil
}

func NewArbitrators(cfg *params.DposConfig) (*arbitrators, error) {
	a := &arbitrators{
		nextCandidates:             make([][]byte, 0),
		accumulativeReward:         0,
		finalRoundChange:           0,
		arbitersRoundReward:        nil,
		illegalBlocksPayloadHashes: make(map[common.Hash]interface{}),
		snapshots:                  make(map[uint64][]*CheckPoint),
		snapshotKeysDesc:           make([]uint64, 0),
		degradation: &degradation{
			inactiveTxs:       make(map[common.Hash]interface{}),
			inactivateHeight:  0,
			understaffedSince: 0,
			state:             DSNormal,
		},
	}
	if err := a.initArbitrators(cfg); err != nil {
		return nil, err
	}
	a.State = NewState(a.GetArbitrators)
	//chainParams.CkpManager.Register(NewCheckpoint(a))
	return a, nil
}