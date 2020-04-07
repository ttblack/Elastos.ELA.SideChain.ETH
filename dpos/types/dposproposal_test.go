package types

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func randomDposProposal() *DPOSProposal {
	var sponsor = make([]byte, len(common.Address{}))
	rand.Read(sponsor)
	blockHash := make([]byte, len(common.Hash{}))
	rand.Read(blockHash)

	sign := make([]byte, 65)
	rand.Read(sign)
	return &DPOSProposal{
		sponsor,
		common.BytesToHash(blockHash),
		0,
		nil,
		nil,
	}
}

func TestDposProposal(t *testing.T) {
	proposal1 := randomDposProposal()
	buf := new(bytes.Buffer)
	err := proposal1.Serialize(buf)
	if err != nil {
		fmt.Println("EncodeRLP error:", err)
	}
	proposal2 := &DPOSProposal{}
	reader := bytes.NewReader(buf.Bytes())
	stream := rlp.NewStream(reader, uint64(reader.Len()))
	proposal2.Deserialize(stream)

	assert.True(t, bytes.Equal(proposal1.Sponsor, proposal2.Sponsor))
	assert.True(t, proposal1.BlockHash == proposal2.BlockHash)
	assert.True(t, 	proposal1.ViewOffset == proposal2.ViewOffset)
	assert.True(t, bytes.Equal(proposal1.Hash().Bytes(), proposal2.Hash().Bytes()))

}