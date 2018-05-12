package blockchain

import (
	"errors"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/keypair"
)

// validate block and associated transactions
func (bc *Blockchain) validateBlock(blk *bcpb.Block, txs []*bcpb.Tx) error {
	// Call the user specified block verifier/validator
	err := bc.bv(blk.Header)
	if err != nil {
		return err
	}

	// Verify required signatures
	if !bc.verifyBlockSignatures(blk) {
		return bcpb.ErrSignatureVerificationFailed
	}

	// Check txs exist in the block
	for i, tid := range blk.Txs {
		if !tid.Equal(txs[i].Digest) {
			return errors.New("tx not in block")
		}
	}

	err = bc.validateTxs(txs)
	if err != nil {
		return err
	}

	// Store transactions
	return bc.tx.SetBatch(txs)
}

// this must be called after the block header has been validated
func (bc *Blockchain) verifyBlockSignatures(blk *bcpb.Block) bool {
	var (
		sh = blk.Header.Hash(bc.h)
		sc int32
	)

	for i := range blk.Header.Signers {
		// Skip unsigned slots
		if len(blk.Signatures[i]) == 0 {
			continue
		}

		kp := keypair.New(bc.curve, bc.h)
		kp.PublicKey = bcpb.PublicKey(blk.Header.Signers[i])
		//log.Printf("%s %s", sh.String(), kp.Address())
		if kp.VerifySignature(sh, blk.Signatures[i]) {
			sc++
		}

	}

	return sc >= blk.Header.S
}

func (bc *Blockchain) validateTxs(txs []*bcpb.Tx) error {
	var err error

	// Validate each tx
	for _, tx := range txs {
		err = bc.validateTx(tx)
		if err != nil {
			break
		}
	}

	return err
}

func (bc *Blockchain) validateTx(tx *bcpb.Tx) error {
	// Validate each tx input
	for _, in := range tx.Inputs {
		//_, err := bc.GetTXO(in)
		_, err := bc.validateTxInput(in)
		if err != nil && err != errBaseTx {
			return err
		}

	}

	return nil
}

// validateTxInput validates the txinput including access authorization and
// signature verification
func (bc *Blockchain) validateTxInput(txi *bcpb.TxInput) (*bcpb.TxOutput, error) {
	if txi.IsBase() {
		return nil, errBaseTx
	}

	txref, err := bc.tx.Get(txi.Ref)
	if err != nil {
		return nil, err
	}

	var (
		txo    = txref.Outputs[txi.Index]
		digest = txi.Hash(bc.h)
		sc     uint8
	)

	// Validate and get number of signatures. This is validated regardless of
	// whether logic is specified
	for i, pk := range txi.PubKeys {
		// Each key must be able to unlock the output
		if !txo.HasPublicKey(pk) {
			return nil, bcpb.ErrNotAuthorized
		}

		// Verify tx input signatures
		kp := keypair.New(bc.curve, bc.h)
		kp.PublicKey = pk
		if kp.VerifySignature(digest, txi.Signatures[i]) {
			sc++
		}

	}

	if len(txo.Logic) == 0 {
		return txo, nil
	}

	// Check required signatures.
	// The first byte in Logic is the required signatures
	reqSigs := uint8(txo.Logic[0])
	if sc < reqSigs {
		return nil, errRequiresMoreSignatures
	}

	return txo, nil
}
