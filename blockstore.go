package blockchain

import (
	"errors"
	"fmt"
	"time"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/hasher"
)

var (
	errInvalidNonce   = errors.New("invalid nonce")
	errHeightMismatch = errors.New("height mismatch")
	// ErrPrevBlockMismatch is returned when a block references a previous
	// block that is not the actual previous block
	ErrPrevBlockMismatch = errors.New("previous block mismatch")
)

// NewGenesisBlock inits everything except the owner
func NewGenesisBlock(h hasher.Hasher) *bcpb.Block {
	return &bcpb.Block{
		Header: &bcpb.BlockHeader{
			Timestamp: time.Now().UnixNano(),
			Nonce:     1,
			PrevBlock: bcpb.NewZeroDigest(h),
		},
	}
}

// BlockStore adds ledger logic around the block storage
type blockStore struct {
	st BlockStorage
}

// SetGenesis sets the genesis block for the blockchain.  This can only be called
// once
func (bc *blockStore) SetGenesis(genesis *bcpb.Block) error {
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

	return store.SetGenesis(gid)
}

// Append verifies and validates the block before appending it to the ledger
func (bc *blockStore) Append(blk *bcpb.Block) (bcpb.Digest, error) {
	if err := bc.checkPrevHeightNonce(blk.Header); err != nil {
		return nil, err
	}

	return bc.st.Add(blk)
}

// checkPrevHeightNonce checks the height, nonce and previous block hash in that
// order.  Height and nonce are checked first as they are cheaper operations.
func (bc *blockStore) checkPrevHeightNonce(blk *bcpb.BlockHeader) (err error) {
	lid, last := bc.st.Last()

	if blk.Height != last.Header.Height+1 {
		// Check height match
		return errHeightMismatch

	} else if blk.Nonce < last.Header.Nonce {
		// New nonce is greater than old one
		return errInvalidNonce

	} else if !lid.Equal(blk.PrevBlock) {
		// Check prev block match
		return ErrPrevBlockMismatch

	}

	return nil
}
