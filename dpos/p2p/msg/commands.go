// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
// 

package msg

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dposp2p"
)

const (
	CmdVersion  = "version"
	CmdVerAck   = "verack"
	CmdAddr     = "addr"
	CmdPing     = "ping"
	CmdPong     = "pong"
	CmdInv      = "inv"
	CmdGetBlock = "getblock"

	CmdReceivedProposal            = "proposal"
	CmdAcceptVote                  = "acc_vote"
	CmdRejectVote                  = "rej_vote"
	CmdGetBlocks                   = "get_blc"
	CmdResponseBlocks              = "res_blc"
	CmdRequestConsensus            = "req_con"
	CmdResponseConsensus           = "res_con"
	CmdRequestProposal             = "req_pro"
	CmdIllegalProposals            = "ill_pro"
	CmdIllegalVotes                = "ill_vote"
	CmdSidechainIllegalData        = "side_ill"
	CmdResponseInactiveArbitrators = "ina_ars"
)

func GetMessageHash(msg dposp2p.Message) common.Hash {
	return common.Hash{}
}
