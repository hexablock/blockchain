package bcpb

import (
	"bytes"
	"encoding/binary"

	"github.com/hexablock/hasher"
)

// NewTxInput creates a new transaction input.  It inits the signatures to the
// the number of given public keys
func NewTxInput(ref Digest, i int32, pubkeys []PublicKey) *TxInput {
	return &TxInput{
		Ref:     ref,
		Index:   i,
		PubKeys: pubkeys,
		// We pre-init this so args can be appended
		Signatures: make([][]byte, len(pubkeys)),
	}
}

// NewBaseTxInput creates a new base TxnInput with arbitrary data as the
// signature.  This is used in for example the genesis block
func NewBaseTxInput(sigData ...[]byte) *TxInput {
	return &TxInput{
		Ref:        nil,
		Index:      -1,
		PubKeys:    nil,
		Signatures: sigData,
	}
}

// IsBase returns true if the reference is not present and the index is set to
// -1.  This is true for creation of entities
func (txi *TxInput) IsBase() bool {
	return txi.Ref == nil && txi.Index == -1
}

// Hash returns a hash of the ref, output index, and public keys
func (txi *TxInput) Hash(hf hasher.Hasher) Digest {
	h := hf.New()

	h.Write(txi.Ref)
	binary.Write(h, binary.BigEndian, txi.Index)

	// Public keys
	for i := range txi.PubKeys {
		h.Write(txi.PubKeys[i])
	}

	// Logic and Args
	args := txi.Args()
	for i := range args {
		h.Write(args[i])
	}

	sh := h.Sum(nil)

	return NewDigest(hf.Name(), sh)
}

// Sign adds the public key and associated signature to the transaction input .
// If returns a ErrNotAuthorized error is the public key is not specified in the
// input
func (txi *TxInput) Sign(pk PublicKey, sig []byte) error {
	i, ok := txi.HasPubKey(pk)
	if !ok {
		return ErrNotAuthorized
	}
	txi.Signatures[i] = sig
	return nil
}

// HasPubKey returns true if the given public key is in the input
func (txi *TxInput) HasPubKey(pubkey []byte) (int, bool) {
	for i := range txi.PubKeys {
		if bytes.Compare(txi.PubKeys[i], pubkey) == 0 {
			return i, true
		}
	}
	return -1, false
}

// AddPubKey adds a public key to the input if it does not exist.  It also
// adjusts the signatures slices accordingly if the addition is successful
func (txi *TxInput) AddPubKey(pk PublicKey) bool {
	if _, ok := txi.HasPubKey(pk); ok {
		return false
	}

	sigs := make([][]byte, len(txi.Signatures)+1)
	copy(sigs[:len(txi.PubKeys)], txi.Signatures[:len(txi.PubKeys)])
	copy(sigs[len(txi.PubKeys)+1:], txi.Signatures[len(txi.PubKeys):])

	txi.PubKeys = append(txi.PubKeys, pk)
	txi.Signatures = sigs

	return true
}

// AddArgs adds inputs args to the txinput. Args are always after all signatures
func (txi *TxInput) AddArgs(args ...[]byte) {
	txi.Signatures = append(txi.Signatures, args...)
}

// Args returns all arguments minus the signatures needed as input. Inputs are
// located after the given number of public keys in the signatures files
func (txi *TxInput) Args() [][]byte {
	return txi.Signatures[len(txi.PubKeys):]
}
