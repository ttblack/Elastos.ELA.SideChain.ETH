package pbft

import (
	"github.com/elastos/Elastos.ELA.SideChain.ETH/chainbridge-core/dpos_msg"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"

	dpeer "github.com/elastos/Elastos.ELA/dpos/p2p/peer"
	"github.com/elastos/Elastos.ELA/events"
	elap2p "github.com/elastos/Elastos.ELA/p2p"
)

func (p *Pbft) SendMsgProposal(proposalMsg elap2p.Message) {
	if p.network == nil {
		panic("direct network is nil")
	}
	p.BroadMessage(proposalMsg)
}

func (p *Pbft) SignData(data []byte) []byte {
	return p.account.Sign(data)
}

func (p *Pbft) GetProducer() []byte {
	return p.account.PublicKeyBytes()
}

func (p *Pbft) OnLayer2Msg(id dpeer.PID, c elap2p.Message) {
	switch c.CMD() {
	case dpos_msg.CmdDepositproposal:
		msg, _ := c.(*dpos_msg.DepositProposalMsg)
		msg.PID = id
		if !p.dispatcher.GetConsensusView().IsProducers(msg.Proposer) {
			log.Error("proposer is not a producer:" + common.Bytes2Hex(msg.Proposer))
			return
		}
		events.Notify(dpos_msg.ETOnProposal, msg)
	}
}