package bcpb

import (
	"bytes"
	"encoding/binary"

	"github.com/hexablock/ledger/types"
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
func (txi *TxInput) Hash(hf types.Hasher) Digest {
	h := hf.New()

	// Ref
	h.Write(txi.Ref)
	// Output index
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

// HasPubKey returns true if the given public key is in the input
func (txi *TxInput) HasPubKey(pubkey []byte) (int, bool) {
	for i := range txi.PubKeys {
		if bytes.Compare(txi.PubKeys[i], pubkey) == 0 {
			return i, true
		}
	}
	return -1, false
}

// AddArgs adds inputs args to the txinput
func (txi *TxInput) AddArgs(args ...[]byte) {
	txi.Signatures = append(txi.Signatures, args...)
}

// Args returns all arguments minus the signatures needed as input. Inputs are
// located after the given number of public keys in the signatures files
func (txi *TxInput) Args() [][]byte {
	return txi.Signatures[len(txi.PubKeys):]
}

// Sign signs the input with the given KeyPair and Hasher. It returns an error
// if the public key is not part of the input or if there is an erro signing
// func (txi *TxInput) Sign(kp keypair.KeyPair, h types.Hasher) error {
// 	// This should match one of the keys defined in the referenced tx's output
// 	i, ok := txi.HasPubKey(kp.PublicKey)
// 	if !ok {
// 		return errors.New("public key not in tx input")
// 	}
//
// 	txsh := txi.Hash(h)
// 	r, s, err := ecdsa.Sign(rand.Reader, &kp.PrivateKey, txsh)
// 	if err == nil {
// 		txi.Signatures[i] = append(r.Bytes(), s.Bytes()...)
// 	}
//
// 	return err
// }

// VerifySignatures verifies all the signatures in the input.  It returns false
// if all signatures are not verified
// func (txi *TxInput) VerifySignatures(curve elliptic.Curve, h types.Hasher) bool {
// 	sh := txi.Hash(h)
// 	var v int
// 	for i := range txi.PubKeys {
// 		if txi.verifySignature(curve, sh, i) {
// 			v++
// 		}
// 	}
//
// 	return v == len(txi.PubKeys)
// }

// verifySignature verifies the signature with the input hash at the given index
// func (txi *TxInput) verifySignature(curve elliptic.Curve, sh []byte, i int) bool {
//
// 	signature := txi.Signatures[i]
// 	pubkey := txi.PubKeys[i]
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

// PubKeyCanUnlockOutput returns true if the the public key will be able to
// unlock the referenced output.  It does so by matching the public key with
// to TxnInput.PubKey.
// func (txi *TxnInput) PubKeyCanUnlockOutput(pk PublicKey) bool {
// 	return txi.PubKey.Equal(pk)
// }
