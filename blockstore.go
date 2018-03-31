package blockchain

import (
	"errors"
	"fmt"
	"time"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/hasher"
)

var (
	errInvalidNonce      = errors.New("invalid nonce")
	errPrevBlockMismatch = errors.New("previous block mismatch")
	errHeightMismatch    = errors.New("height mismatch")
)

// NewGenesisBlock inits everything except the owner
func NewGenesisBlock(txs []*bcpb.Tx, h hasher.Hasher) *bcpb.Block {
	blk := &bcpb.Block{
		Header: &bcpb.BlockHeader{
			Timestamp: time.Now().UnixNano(),
			Nonce:     1,
			PrevBlock: bcpb.NewZeroDigest(h),
		},
	}

	blk.SetTxs(txs, h)

	return blk
}

// BlockStore adds ledger logic around the block storage
type BlockStore struct {
	st BlockStorage
}

// SetGenesis sets the genesis block for the blockchain.  This can only be called
// once
func (bc *BlockStore) SetGenesis(genesis *bcpb.Block) error {
	store := bc.st

	// Check if we already have a genesis block
	if _, gen := store.Genesis(); gen != nil {
		return fmt.Errorf("genesis block already set")
	}

	// This will return an error if it already exists
	gid, err := store.Add(genesis)
	if err != nil {
		return err
	}

	// Update all references to genesis block
	if err = store.SetGenesis(gid); err != nil {
		return err
	}
	if err = store.SetLast(gid); err != nil {
		return err
	}
	if err = store.SetLastExec(gid); err != nil {
		return err
	}

	return nil
}

// Append verifies and validates the block before appending it to the ledger
func (bc *BlockStore) Append(blk *bcpb.Block) (bcpb.Digest, error) {
	if err := bc.checkPrevHeightNonce(blk.Header); err != nil {
		return nil, err
	}

	return bc.st.Add(blk)
}

// NextBlock returns the next block in the ledger.  This is then ratified and
// submitted to be added to the ledger
func (bc *BlockStore) NextBlock() *bcpb.Block {
	lid, last := bc.st.Last()

	blk := bcpb.NewBlock()
	blk.Header = &bcpb.BlockHeader{
		Height:    last.Header.Height + 1,
		PrevBlock: lid,
		Timestamp: time.Now().UnixNano(),
		Nonce:     last.Header.Nonce + 1,
	}

	return blk
}

func (bc *BlockStore) checkPrevHeightNonce(blk *bcpb.BlockHeader) (err error) {
	lid, last := bc.st.Last()

	if blk.Height != last.Header.Height+1 {
		// Check height match
		return errHeightMismatch

	} else if !lid.Equal(blk.PrevBlock) {
		// Check prev block match
		return errPrevBlockMismatch

	} else if blk.Nonce < last.Header.Nonce {
		// New nonce is greater than old one
		return errInvalidNonce

	}

	return nil
}
