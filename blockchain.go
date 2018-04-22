package blockchain

import (
	"crypto/elliptic"
	"errors"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/stores"
	"github.com/hexablock/hasher"
)

var (
	errBaseTx                 = errors.New("base transaction")
	errRequiresMoreSignatures = errors.New("requires more signatures")
)

// BlockValidator is the validator function called to validate a block before
// signatures a verfified
type BlockValidator func(*bcpb.BlockHeader) error

// TxInputOutputValidator validates a single input.  It takes the referenced
// output and the associated input as arguments. TxOutput <- TxInput
type TxInputOutputValidator func(ref *bcpb.TxOutput, in *bcpb.TxInput) error

// BlockStorage implements a store for ledger blocks
type BlockStorage interface {
	// Get a block by its digest id
	Get(bcpb.Digest) (*bcpb.Block, error)
	// Returns the genesis  block
	Genesis() (bcpb.Digest, *bcpb.Block)
	// Last block in the ledger
	Last() (bcpb.Digest, *bcpb.Block)
	// Get last executed
	LastExec() (bcpb.Digest, *bcpb.Block)
	// Sets the genesis block digest. It assumes the actual block is already in
	// in the store
	SetGenesis(bcpb.Digest) error
	// Set last block digest. It assumes the actual block is already in
	// in the store
	SetLast(bcpb.Digest) error
	// Sets the last executed block. It assumes the actual block is already in
	// in the store
	SetLastExec(bcpb.Digest) error
	// Returns true if the block by the given digest exists
	Exists(bcpb.Digest) bool
	// Adds a block to the ledger returning an error if it already exists
	Add(*bcpb.Block) (bcpb.Digest, error)
	// Iter iterates of each block in the ledger
	Iter(f stores.BlockIterator) error
}

// TxStorage implements a transaction store
type TxStorage interface {
	// Get a transaction
	Get(bcpb.Digest) (*bcpb.Tx, error)
	// Set a transaction
	Set(*bcpb.Tx) error
	// Set a batch of transactions
	SetBatch([]*bcpb.Tx) error
	// Iterate over all transactions
	Iter(func(bcpb.Tx) error)
}

// DataKeyIndex is an index of DataKey to the txref and output index of all
// unspent outputs.
type DataKeyIndex interface {
	Get(key bcpb.DataKey) (bcpb.Digest, int32, error)
	Set(key bcpb.DataKey, ref bcpb.Digest, idx int32) error
	Iter(prefix bcpb.DataKey, iter stores.DataKeyIterator) error
}

// Blockchain is a blockchain instance that is able to perform all verification
// but does not include the consensus logic
type Blockchain struct {
	h     hasher.Hasher
	curve elliptic.Curve
	// Block validation function
	bv BlockValidator

	blk *blockStore
	tx  *txStore
}

// New instantiates a new blockchain.  By default block validation is disabled
func New(conf *Config) *Blockchain {
	return &Blockchain{
		// Hash function
		h: conf.Hasher,
		// Elliptic curve
		curve: conf.Curve,
		// Disable block validation
		bv: func(*bcpb.BlockHeader) error { return nil },
		// Block store
		blk: &blockStore{conf.BlockStorage},
		// Tx store
		tx: &txStore{conf.TxStorage, conf.DataKeyIndex},
	}
}

// SetBlockValidator sets the block validator function
func (bc *Blockchain) SetBlockValidator(bv BlockValidator) {
	bc.bv = bv
}

// Hasher returns the configured hash function used by the block chain.
func (bc *Blockchain) Hasher() hasher.Hasher {
	return bc.h
}

// NewTxInput returns a new TxInput for the given key to use in a tx
func (bc *Blockchain) NewTxInput(key bcpb.DataKey) (*bcpb.TxInput, error) {
	return bc.tx.NewTxInput(key)
}

// Genesis returns the genesis block or nil if the blockchain has not been
// initialized
func (bc *Blockchain) Genesis() *bcpb.Block {
	_, b := bc.blk.st.Genesis()
	return b
}

// Last returns the last commited block in the chain
func (bc *Blockchain) Last() *bcpb.Block {
	_, last := bc.blk.st.Last()
	return last
}

// SetGenesis sets the genesis block and the associated transactions
func (bc *Blockchain) SetGenesis(genesis *bcpb.Block, txs []*bcpb.Tx) error {
	err := bc.validateBlock(genesis, txs)
	if err != nil {
		return err
	}

	err = bc.blk.SetGenesis(genesis)
	if err == nil {
		// If we succeed we set the last block digest to the zero hash
		err = bc.blk.st.SetLast(bcpb.NewZeroDigest(bc.h))
	}

	return err
}

// SetLastExec marks the given digest as the last executed block
func (bc *Blockchain) SetLastExec(digest bcpb.Digest) error {
	return bc.blk.st.SetLastExec(digest)
}

// Append appends the block and txs to the ledger.  The supplied transactions
// must be part of the block.  This does not update the last block reference or
// index any of the txos
func (bc *Blockchain) Append(blk *bcpb.Block, txs []*bcpb.Tx) (bcpb.Digest, error) {
	err := bc.validateBlock(blk, txs)
	if err == nil {
		return bc.blk.Append(blk)
	}
	return nil, err
}

// Commit commits the block given by the id. It ensures it is the next in line
// i.e. the previous hash matches the current last block, sets the last block
// to the given id and indexes all transaction outputs in the block
func (bc *Blockchain) Commit(id bcpb.Digest) error {
	// Get stored block thats being committed
	blk, err := bc.blk.st.Get(id)
	if err != nil {
		return err
	}

	// Ensure the blocks previous hash matches that of the last block
	lid, _ := bc.blk.st.Last()
	if !blk.Header.PrevBlock.Equal(lid) {
		return ErrPrevBlockMismatch
	}

	// Set the given id as the last block
	err = bc.blk.st.SetLast(id)
	if err == nil {
		// Index the tx outputs
		err = bc.indexTxos(blk)
	}

	return err
}

// GetTXO returns the txo referenced by the TxInput. It returns an error
// if access is not authorized or any validation fails
func (bc *Blockchain) GetTXO(txi *bcpb.TxInput) (*bcpb.TxOutput, error) {
	return bc.validateTxInput(txi)
}

// GetTXOByDataKey returns the TxOutput for the given key.  It is the DataKey's
// last state
func (bc *Blockchain) GetTXOByDataKey(key bcpb.DataKey) (*bcpb.TxOutput, error) {
	tx, i, err := bc.tx.GetDataKeyTx(key)
	if err != nil {
		return nil, err
	}

	return tx.Outputs[i], nil
}

func (bc *Blockchain) indexTxos(blk *bcpb.Block) (err error) {
	// Get txs in the block
	txs := make([]*bcpb.Tx, len(blk.Txs))
	for i, tid := range blk.Txs {
		txs[i], err = bc.tx.tx.Get(tid)
		if err != nil {
			return err
		}
	}

	// Index the tx outputs
	return bc.tx.indexTxos(txs)
}
