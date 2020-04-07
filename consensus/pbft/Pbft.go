package pbft

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/accounts"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/dpos"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/node"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/params"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rlp"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/rpc"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/spv"
	lru "github.com/hashicorp/golang-lru"
	"io"
	"math/big"
	"sync"
	"time"

	"github.com/elastos/Elastos.ELA.SideChain.ETH/common"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/consensus"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/consensus/clique"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/state"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/core/types"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/crypto"
	"github.com/elastos/Elastos.ELA.SideChain.ETH/log"
)

var (
	extraVanity = 32                     // Fixed number of extra-data prefix bytes reserved for signer vanity
	extraSeal   = crypto.SignatureLength // Fixed number of extra-data suffix bytes reserved for signer seal
	extraElaHeight = 8
	uncleHash = types.CalcUncleHash(nil) // Always Keccak256(RLP([])) as uncles are meaningless outside of PoW.
	inmemorySignatures = 4096
)

var (
	errMissingSignature = errors.New("extra-data 65 byte signature suffix missing")
	// errUnknownBlock is returned when the list of signers is requested for a block
	// that is not part of the local blockchain.
	errUnknownBlock = errors.New("unknown block")

	errUnauthorizedSigner = errors.New("unauthorized producer")

	errNotOnDuty = errors.New("producer is not on duty")
)

type PBFT struct {
	lock   sync.RWMutex   // Protects the signer fields
	signatures *lru.ARCCache // Signatures of recent blocks to speed up mining
	signer common.Address // Ethereum address of the signing key
	signFn clique.SignerFn       // Signer function to authorize hashes with
	config *params.DposConfig
	arbiter *dpos.Arbitrator
	ctx *node.ServiceContext

	running bool
}

func New(config *params.DposConfig, ctx *node.ServiceContext) *PBFT {
	// Set any missing consensus parameters to their defaults
	conf := *config
	// Allocate the snapshot caches and create the engine
	signatures, _ := lru.NewARC(inmemorySignatures)

	return &PBFT{
		config:     &conf,
		ctx:     	ctx,
		signatures: signatures,
		running: false,
	}
}

func  (c *PBFT) GetDposConfig() *params.DposConfig {
	return c.config
}

func (c *PBFT) GetDposStorePath() string {
	return c.ctx.ResolvePath("dpos")
}

func (c *PBFT) GetDataDir() string {
	return c.ctx.ResolvePath("")
}

func (c *PBFT) GetSignFn() clique.SignerFn {
	return c.signFn
}

func (c *PBFT) SetArbiter(arbiter *dpos.Arbitrator) {
	c.arbiter = arbiter
}

// Author implements consensus.Engine, returning the Ethereum address recovered
// from the signature in the header's extra-data section.
func (c *PBFT) Author(header *types.Header) (common.Address, error) {
	return ecrecover(header, c.signatures)
}

// ecrecover extracts the Ethereum account address from a signed header.
func ecrecover(header *types.Header, sigcache *lru.ARCCache) (common.Address, error) {
	// If the signature's already cached, return that
	hash := header.Hash()
	if address, known := sigcache.Get(hash); known {
		return address.(common.Address), nil
	}
	// Retrieve the signature from the header extra-data
	if len(header.Extra) < extraSeal {
		return common.Address{}, errMissingSignature
	}
	signature := header.Extra[len(header.Extra)-extraSeal:]

	// Recover the public key and the Ethereum address
	pubkey, err := crypto.Ecrecover(clique.SealHash(header).Bytes(), signature)
	if err != nil {
		return common.Address{}, err
	}
	var signer common.Address
	copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

	sigcache.Add(hash, signer)
	return signer, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *PBFT) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	return c.verifyHeader(chain, header, nil)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *PBFT) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	abort := make(chan struct{})
	results := make(chan error, len(headers))

	go func() {
		for i, header := range headers {
			err := c.verifyHeader(chain, header, headers[:i])

			select {
			case <-abort:
				return
			case results <- err:
			}
		}
	}()
	return abort, results
}

// verifyHeader checks whether a header conforms to the consensus rules.The
// caller may optionally pass in a batch of parents (ascending order) to avoid
// looking those up from the database. This is useful for concurrently verifying
// a batch of new headers.
func (c *PBFT) verifyHeader(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	if header.Number == nil {
		return errUnknownBlock
	}

	return nil
}

// VerifyUncles implements consensus.Engine, always returning an error for any
// uncles as this consensus mechanism doesn't permit uncles.
func (c *PBFT) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if len(block.Uncles()) > 0 {
		return errors.New("uncles not allowed")
	}
	return nil
}

// VerifySeal implements consensus.Engine, checking whether the signature contained
// in the header satisfies the consensus protocol requirements.
func (c *PBFT) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return c.verifySeal(chain, header, nil)
}

// verifySeal checks whether the signature contained in the header satisfies the
// consensus protocol requirements. The method accepts an optional list of parent
// headers that aren't yet part of the local blockchain to generate the snapshots
// from.
func (c *PBFT) verifySeal(chain consensus.ChainReader, header *types.Header, parents []*types.Header) error {
	// Verifying the genesis block is not supported
	return nil
}

// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *PBFT) Prepare(chain consensus.ChainReader, header *types.Header) error {
	// If the block isn't a checkpoint, cast a random vote (good enough for now)
	header.Coinbase = common.Address{}
	header.Nonce = types.BlockNonce{}
	header.Time = uint64(time.Now().Unix())
	number := header.Number.Uint64()

	// Ensure the extra data has all its components
	if len(header.Extra) < extraVanity {
		header.Extra = append(header.Extra, bytes.Repeat([]byte{0x00}, extraVanity-len(header.Extra))...)
	}
	header.Extra = header.Extra[:extraVanity]
	header.Extra = append(header.Extra, make([]byte, extraElaHeight)...)
	header.Extra = append(header.Extra, make([]byte, extraSeal)...)

	// Mix digest is reserved for now, set to empty
	header.MixDigest = common.Hash{}

	// Ensure the timestamp has the correct delay
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}

	if c.running == false && c.arbiter != nil {
		c.Start()
	}
	return nil
}

// Finalize implements consensus.Engine, ensuring no uncles are set, nor block
// rewards given.
func (c *PBFT) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header) {
	// No block rewards in PBFT, so the state remains as is and uncles are dropped
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)
}

// FinalizeAndAssemble implements consensus.Engine, ensuring no uncles are set,
// nor block rewards given, and returns the final block.
func (c *PBFT) FinalizeAndAssemble(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// No block rewards in PoA, so the state remains as is and uncles are dropped
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	header.UncleHash = types.CalcUncleHash(nil)

	// Assemble and return the final block for sealing
	block := types.NewBlock(header, txs, nil, receipts)
	if (c.running) {
		c.arbiter.OnBlockReceived(block, false)
	}
	return block, nil
}

// Authorize injects a private key into the consensus engine to mint new blocks
// with.
func (c *PBFT) Authorize(signer common.Address, signFn clique.SignerFn) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.signer = signer
	c.signFn = signFn
}

// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *PBFT) Seal(chain consensus.ChainReader, block *types.Block, results chan<- *types.Block, stop <-chan struct{}) error {
	if chain.IsDangerChain() {
		err := errors.New("too many evil signers on the chain")
		log.Error(err.Error())
		return nil
	}

	if !c.arbiter.IsArbiter() {
		return errUnauthorizedSigner
	}

	if !c.arbiter.IsOnDuty() {
		return errNotOnDuty
	}

	header := block.Header()

	length := len(header.Extra)
	in := new(bytes.Buffer)
	elaHeight := spv.GetElaHeight()
	if err := binary.Write(in, binary.BigEndian, elaHeight); err == nil {
		copy(header.Extra[length-extraElaHeight-extraSeal:length-extraSeal], in.Bytes())
	}

	c.lock.RLock()
	signer, signFn := c.signer, c.signFn
	c.lock.RUnlock()

	// Sign all the things!
	sighash, err := signFn(accounts.Account{Address: signer}, accounts.MimetypeClique, PbftRLP(header))
	if err != nil {
		return err
	}
	copy(header.Extra[length - extraSeal:], sighash)
	go func() {
		select {
		case results <- block.WithSeal(header):
		default:
			log.Warn("Sealing result is not read by miner", "sealhash", c.SealHash(header))
		}
	}()

	return nil
}

func PbftRLP(header *types.Header) []byte {
	b := new(bytes.Buffer)
	encodeSigHeader(b, header)
	return b.Bytes()
}

func encodeSigHeader(w io.Writer, header *types.Header) {
	err := rlp.Encode(w, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-crypto.SignatureLength], // Yes, this will panic if extra is too short
		header.MixDigest,
		header.Nonce,
	})
	if err != nil {
		panic("can't encode: " + err.Error())
	}
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have based on the previous blocks in the chain and the
// current signer.
func (c *PBFT) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	return big.NewInt(0)
}


// SealHash returns the hash of a block prior to it being sealed.
func (c *PBFT) SealHash(header *types.Header) common.Hash {
	return clique.SealHash(header)
}

func (c *PBFT) Start() error {
	if c.running == true {
		return errors.New("PBFT is running")
	}
	c.arbiter.Start()
	c.running = true
	return nil
}

// Close implements consensus.Engine. It's a noop for clique as there are no background threads.
func (c *PBFT) Close() error {
	c.arbiter.Stop()
	c.running = false
	return nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *PBFT) APIs(chain consensus.ChainReader) []rpc.API {
	return []rpc.API{{
		Namespace: "pbft",
		Version:   "1.0",
		Service:   &API{chain: chain, engine: c},
		Public:    false,
	}}
}

func (c *PBFT) SignersCount() int {
	return c.arbiter.GetProducerCount()
}

