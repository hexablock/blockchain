package blockchain

import (
	"errors"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/keypair"
)

// validate block and associated transactions
func (bc *Blockchain) validateBlock(blk *bcpb.Block, txs []*bcpb.Tx) error {
	// Validate block
	if err := bc.bv(blk); err != nil {
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
		if kp.VerifySignature(sh, blk.Signatures[i]) {
			sc++
		}

	}

	return sc >= blk.Header.S
}

// Validate validates the N, S, Q values as well as the ProposerIndex
// func (bc *Blockchain) validateBlockHeader(header *bcpb.BlockHeader) error {
// 	if int32(len(header.Signers)) != header.N {
// 		return errors.New("not enough signers")
// 	}
//
// 	if header.ProposerIndex < int32(len(header.Signers)) {
// 		return errors.New("invalid proposer index")
// 	}
//
// 	q := (header.N / 2) + 1
// 	if header.S < q {
// 		return errors.New("invalid number of required signatures")
// 	}
// 	if header.Q < q {
// 		return errors.New("invalid number of required commits")
// 	}
//
// 	return nil
// }
