package types

import (
	"bytes"
	"fmt"
	"testing"

	"crypto/rand"
	"github.com/stretchr/testify/assert"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
)

func randomDposProposalVote() *DPOSProposalVote {
	hash := make([]byte, len(common.Hash{}))
	rand.Read(hash)

	var signer = make([]byte, len(common.Address{}))
	rand.Read(signer)

	var sign = make([]byte, 65)
	rand.Read(sign)

	return &DPOSProposalVote{
		common.BytesToHash(hash),
		signer,
		true,
		sign,
		nil,
	}
}

func TestDposProposalVote(t *testing.T) {
	vote1 := randomDposProposalVote()
	buf := new(bytes.Buffer)
	err := vote1.Serialize(buf)
	if err != nil {
		fmt.Println("EncodeRLP error:", err)
	}
	vote2 := &DPOSProposalVote{}
	reader := bytes.NewReader(buf.Bytes())
	stream := rlp.NewStream(reader, uint64(reader.Len()))
	vote2.Deserialize(stream)

	assert.True(t, bytes.Equal(vote1.ProposalHash.Bytes(), vote2.ProposalHash.Bytes()))
	assert.True(t, bytes.Equal(vote1.Signer, vote2.Signer))
	assert.True(t, vote1.Accept == vote2.Accept)
	assert.True(t, bytes.Equal(vote1.Sign, vote2.Sign))
	assert.True(t, bytes.Equal(vote1.Hash().Bytes(), vote2.Hash().Bytes()))


}