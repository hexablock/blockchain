package bcpb

import (
	"encoding/binary"
	"errors"
	"log"
	"math"
	"time"

	"github.com/hexablock/hasher"
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
	// ErrSignerAlreadySigned is returned when a signer tries to sign when they
	// already have
	ErrSignerAlreadySigned = errors.New("signer already signed")
)

// Proposer returns the public key of the block proposer
func (header *BlockHeader) Proposer() PublicKey {
	return header.Signers[header.ProposerIndex]
}

// HasSigners returns true if the block has the number of signers as specified
// by N
func (header *BlockHeader) HasSigners() bool {
	return int(header.N) == len(header.Signers)
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
	binary.Write(hf, binary.BigEndian, header.S)
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

// SetHash sets the block hash using the given hash function
func (blk *Block) SetHash(h hasher.Hasher) {
	blk.Header.Root, _ = Digests(blk.Txs).Root()
	blk.Digest = blk.Header.Hash(h)
}

// SignatureCount returns the number of signatures in the block.  This is
// different from the S value which represents the 'required' signatures
func (blk *Block) SignatureCount() int32 {
	if len(blk.Signatures) == 0 {
		return 0
	}

	var c int32
	for i := range blk.Signatures {
		if len(blk.Signatures[i]) > 0 {
			c++
		}
	}
	return c
}

// HasSignatures returns true if the block has the required number of signatures
// as specified by the S value in the header
func (blk *Block) HasSignatures() bool {
	return blk.SignatureCount() == blk.Header.S
}

// ProposerSigned returns true if the proposer has signed the block
func (blk *Block) ProposerSigned() bool {
	return len(blk.Signatures[blk.Header.ProposerIndex]) > 0
}

// SetProposer sets the proposer index given the public key.  If the public is
// not in the signers it is added and then the index is set
func (blk *Block) SetProposer(pubkey PublicKey) {
	i := blk.Header.SignerIndex(pubkey)
	if i >= 0 {
		blk.Header.ProposerIndex = int32(i)
		return
	}

	// Add and update index
	blk.AddSigner(pubkey)
	blk.Header.ProposerIndex = int32(len(blk.Header.Signers) - 1)
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

	if len(blk.Signatures[i]) != 0 {
		log.Printf("FAILURE %x %x", blk.Header.Signers[i][:8], blk.Signatures[i][:8])
		return ErrSignerAlreadySigned
	}

	blk.Signatures[i] = signature
	return nil
}

// SetTxs sets the transaction id's to the block
func (blk *Block) SetTxs(txns []*Tx, hf hasher.Hasher) {
	blk.Txs = make([]Digest, len(txns))

	for i, tx := range txns {
		if len(tx.Digest) == 0 {
			tx.SetDigest(hf)
		}

		blk.Txs[i] = tx.Digest.Copy()
	}
}

// Clone clones the block
func (blk *Block) Clone() *Block {
	hdr := *blk.Header

	b := &Block{
		Header:     &hdr,
		Txs:        make([]Digest, len(blk.Txs)),
		Signatures: make([][]byte, len(blk.Signatures)),
		Digest:     blk.Digest.Copy(),
	}
	copy(b.Txs, blk.Txs)
	copy(b.Signatures, blk.Signatures)

	return b
}
