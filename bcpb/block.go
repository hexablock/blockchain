package bcpb

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/hexablock/blockchain/hasher"
)

const nonceMax = math.MaxUint64

var (
	// ErrSignatureVerificationFailed is returned when a signature cannot be
	// verified
	ErrSignatureVerificationFailed = errors.New("signature verification failed")
	// ErrNotAuthorized is returned when the given public is not allowed to
	// access the output
	ErrNotAuthorized = errors.New("not authorized")
	// ErrSignerNotInBlock is returned when a signer signs a block but is not
	// a participant block signer
	ErrSignerNotInBlock = errors.New("signer not in block")
)

// Proposer returns the public key of the block proposer
func (header *BlockHeader) Proposer() PublicKey {
	return header.Signers[header.ProposerIndex]
}

// SignerIndex returns the index in the signers list given the public key
func (header *BlockHeader) SignerIndex(pubkey PublicKey) int {
	for i, h := range header.Signers {
		if h.Equal(pubkey) {
			return i
		}
	}
	return -1
}

// Hash returns the hash of the header
func (header *BlockHeader) Hash(h hasher.Hasher) Digest {
	hf := h.New()

	binary.Write(hf, binary.BigEndian, header.Height)
	hf.Write(header.PrevBlock)
	binary.Write(hf, binary.BigEndian, header.Timestamp)
	binary.Write(hf, binary.BigEndian, header.Nonce)
	hf.Write(header.Root)

	for i := range header.Signers {
		hf.Write(header.Signers[i])
	}

	binary.Write(hf, binary.BigEndian, header.ProposerIndex)
	binary.Write(hf, binary.BigEndian, header.N)
	binary.Write(hf, binary.BigEndian, header.Q)

	sh := hf.Sum(nil)

	return NewDigest(h.Name(), sh)
}

// NewBlock returns a new empty block
func NewBlock() *Block {
	return &Block{
		Header: &BlockHeader{
			Timestamp: time.Now().UnixNano(),
			Signers:   make([]PublicKey, 0),
		},
		Txs:        make([]Digest, 0),
		Signatures: make([][]byte, 0),
	}
}

// Height returns the block height
func (blk *Block) Height() uint32 {
	return blk.Header.Height
}

// AddSigner adds a signer i.e public key to the block header
func (blk *Block) AddSigner(pubkey PublicKey) {
	blk.Header.Signers = append(blk.Header.Signers, pubkey)

	a := make([][]byte, len(blk.Header.Signers))
	copy(a, blk.Signatures)

	blk.Signatures = a
}

// SetSigners sets the signers resetting all signatures
func (blk *Block) SetSigners(signers ...PublicKey) {

	blk.Header.Signers = make([]PublicKey, len(signers))
	copy(blk.Header.Signers, signers)

	blk.Signatures = make([][]byte, len(blk.Header.Signers))
}

// Sign adds the public key and signature to the block.  They are added at the
// same index.  If the public key is already in Signers then the signature is
// set to the corresponding index in the Signatures field
func (blk *Block) Sign(pubkey PublicKey, signature []byte) error {
	i := blk.Header.SignerIndex(pubkey)
	// Signer not in block
	if i < 0 {
		return ErrSignerNotInBlock
	}

	blk.Signatures[i] = signature
	return nil
}

// Sign generates and returns the signature using the keypair
// func (blk *Block) Sign(kp keypair.KeyPair, h types.Hasher) ([]byte, error) {
//
// 	sh := blk.Header.Hash(h)
// 	r, s, err := ecdsa.Sign(rand.Reader, &kp.PrivateKey, sh)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return append(r.Bytes(), s.Bytes()...), nil
//
// }

// func (blk *Block) VerifySignatures(curve elliptic.Curve, h types.Hasher) bool {
// 	sh := blk.Header.Hash(h)
//
// 	var sc int32
//
// 	for i := range blk.Header.Signers {
// 		// Skip unsigned slots
// 		if len(blk.Signatures[i]) == 0 {
// 			continue
// 		}
//
// 		//
// 		// 		if !verifySignature(curve, blk.Header.Signers[i], blk.Signatures[i], sh) {
// 		// 			return false
// 		// 		}
// 		//
// 		// 		// Update verfied signature count
// 		// 		sc++
// 	}
//
// 	return sc >= blk.Header.S
// }

// SetTxs sets the transaction id's to the block
func (blk *Block) SetTxs(txns []*Tx, hf hasher.Hasher) {
	blk.Txs = make([]Digest, len(txns))

	for i, tx := range txns {
		if tx.Digest == nil {
			tx.SetDigest(hf)
		}

		blk.Txs[i] = tx.Digest.Copy()
	}

	blk.Header.Root = rootHash(blk.Txs, hf)
}

// Clone clones the block
func (blk *Block) Clone() *Block {
	hdr := *blk.Header

	b := &Block{
		Header:     &hdr,
		Txs:        make([]Digest, len(blk.Txs)),
		Signatures: make([][]byte, len(blk.Signatures)),
	}
	copy(b.Txs, blk.Txs)
	copy(b.Signatures, blk.Signatures)

	return b
}

// func verifySignature(curve elliptic.Curve, pubkey PublicKey, signature []byte, sh []byte) bool {
//
// 	r := big.Int{}
// 	s := big.Int{}
// 	sigLen := len(signature)
// 	r.SetBytes(signature[:(sigLen / 2)])
// 	s.SetBytes(signature[(sigLen / 2):])
//
// 	x := big.Int{}
// 	y := big.Int{}
// 	keyLen := len(pubkey)
// 	x.SetBytes(pubkey[:(keyLen / 2)])
// 	y.SetBytes(pubkey[(keyLen / 2):])
//
// 	rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
// 	return ecdsa.Verify(&rawPubKey, sh, &r, &s)
// }

func rootHash(list []Digest, h hasher.Hasher) Digest {
	hf := h.New()
	for _, l := range list {
		hf.Write(l)
	}
	sh := hf.Sum(nil)
	return NewDigest(h.Name(), sh)
}
