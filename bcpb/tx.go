package bcpb

import (
	"encoding/binary"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/hexablock/hasher"
)

// Hash returns the hash digest of the header
func (header *TxHeader) Hash(h hasher.Hasher) Digest {
	hf := h.New()

	binary.Write(hf, binary.BigEndian, header.Timestamp)
	hf.Write(header.Data)
	binary.Write(hf, binary.BigEndian, header.DataSize)

	sh := hf.Sum(nil)
	return NewDigest(h.Name(), sh)
}

// NewTx returns a new transaction
func NewTx() *Tx {
	return &Tx{
		Header: &TxHeader{
			Timestamp: time.Now().UnixNano(),
		},
		Inputs:  make([]*TxInput, 0),
		Outputs: make([]*TxOutput, 0),
	}
}

// NewBaseTx returns a new base transaction.  This is used when new entities
// are being created
func NewBaseTx(pubkeys ...PublicKey) *Tx {
	tx := NewTx()
	tx.AddInput(NewTxInput(nil, -1, pubkeys))
	return tx
}

// AddInput appends the given input to the transaction inputs
func (tx *Tx) AddInput(in *TxInput) {
	tx.Inputs = append(tx.Inputs, in)
}

// AddOutput appends the given output to the transaction outputs
func (tx *Tx) AddOutput(out *TxOutput) {
	tx.Outputs = append(tx.Outputs, out)
}

// DataHash hashes all inputs and outputs updating the DataSize in the header
// and returning the data digest
func (tx *Tx) DataHash(h hasher.Hasher) Digest {
	var (
		hf = h.New()
		s  int64
	)

	for i := range tx.Inputs {
		b, _ := proto.Marshal(tx.Inputs[i])
		hf.Write(b)
		s += int64(len(b))
	}
	for i := range tx.Outputs {
		b, _ := proto.Marshal(tx.Outputs[i])
		hf.Write(b)
		s += int64(len(b))
	}

	// Set DataLength in header
	tx.Header.DataSize = s

	sh := hf.Sum(nil)

	return NewDigest(h.Name(), sh)
}

// SetDigest computes the hash of the tx and set the digest field
func (tx *Tx) SetDigest(h hasher.Hasher) {
	tx.Header.Data = tx.DataHash(h)
	tx.Digest = tx.Header.Hash(h)
}

// IsBase returns true if this is a base tx i.e. inputs do not reference any
// outputs
func (tx *Tx) IsBase() bool {
	return tx.Inputs[0].IsBase()
}

// HasPublicKey return true if the public key is one of the public keys
// listed in the output
func (txo *TxOutput) HasPublicKey(pk PublicKey) bool {
	if len(txo.PubKeys) == 0 {
		return true
	}

	for _, p := range txo.PubKeys {
		if p.Equal(p) {
			return true
		}
	}

	return false
}

// RemovePublicKey removes the public key returning true if it was removed
func (txo *TxOutput) RemovePublicKey(pk PublicKey) bool {
	for i := range txo.PubKeys {
		if txo.PubKeys[i].Equal(pk) {
			txo.PubKeys = append(txo.PubKeys[0:i], txo.PubKeys[i+1:]...)
			return true
		}
	}
	return false
}

// SetRequiredSignatures sets the required signatures to mutate the output
func (txo *TxOutput) SetRequiredSignatures(c uint8) {
	if len(txo.Logic) > 0 {
		txo.Logic[0] = c
	} else {
		txo.Logic = []byte{c}
	}
}

// Copy returns a copy of the transaction output
func (txo *TxOutput) Copy() *TxOutput {
	o := &TxOutput{
		DataKey: txo.DataKey,
		PubKeys: make([]PublicKey, len(txo.PubKeys)),
		Counter: txo.Counter,
		Logic:   make([]byte, len(txo.Logic)),
	}

	for i := range txo.PubKeys {
		o.PubKeys[i] = txo.PubKeys[i]
	}
	copy(o.Logic, txo.Logic)

	return o
}
