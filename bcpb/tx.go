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
