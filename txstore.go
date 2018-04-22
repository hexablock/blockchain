package blockchain

import (
	"errors"

	"github.com/hexablock/blockchain/bcpb"
)

// TxStore adds ledger logic around the store
type TxStore struct {
	tx  TxStorage
	dki DataKeyIndex
}

// SetBatch validates the transaction are not spent before setting them to the
// store
func (st *TxStore) SetBatch(txs []*bcpb.Tx) error {
	unspent := st.FindUnspent()

	for _, tx := range txs {

		// Make sure inputs haven't been spent
		for _, in := range tx.Inputs {
			if in.IsBase() {
				continue
			}

			_, ok := unspent[in.Ref.String()]
			if !ok {
				return errors.New("tx already spent")
			}

		}
	}

	return st.tx.SetBatch(txs)
}

// FindUTX finds transaction with unused outputs for the given public key
func (st *TxStore) FindUTX(pubkey bcpb.PublicKey) map[string]bcpb.Tx {
	// Get all unspent
	unspent := st.FindUnspent()
	// Filter by public key
	for k, v := range unspent {
		for _, out := range v.Outputs {
			if !out.HasPublicKey(pubkey) {
				delete(unspent, k)
				break
			}
		}
	}

	return unspent
}

// FindUnspent finds all transactions whose outputs are not references by any
// inputs
func (st *TxStore) FindUnspent() map[string]bcpb.Tx {
	unspent := make(map[string]bcpb.Tx)
	spent := make(map[string]struct{})

	st.tx.Iter(func(tx bcpb.Tx) error {
		// Check if its already marked as spent
		if _, ok := spent[tx.Digest.String()]; !ok {
			// Mark as unspent
			unspent[tx.Digest.String()] = tx
		}

		if tx.IsBase() {
			return nil
		}

		// Get through each input and remove any referenced tx
		for _, in := range tx.Inputs {
			tid := in.Ref.String()
			// Mark referenced tx as spent
			spent[tid] = struct{}{}

			// Remove from unspent in case it marked so
			if _, ok := unspent[tid]; ok {
				delete(unspent, tid)
			}
		}

		return nil
	})

	return unspent
}

// GetDataKeyTx returns the last transaction and output index associated to the
// DataKey.  This is the latest state of the data key
func (st *TxStore) GetDataKeyTx(key bcpb.DataKey) (*bcpb.Tx, int32, error) {
	ref, i, err := st.dki.Get(key)
	if err != nil {
		return nil, -1, err
	}

	tx, err := st.tx.Get(ref)
	return tx, i, err
}

// NewTxInput returns a new TxInput for the DataKey.  This is used to contruct
// transaction inputs
func (st *TxStore) NewTxInput(key bcpb.DataKey) (*bcpb.TxInput, error) {
	ref, i, err := st.dki.Get(key)
	if err != nil {
		return nil, err
	}

	tx, err := st.tx.Get(ref)
	if err != nil {
		return nil, err
	}

	return bcpb.NewTxInput(ref, i, tx.Outputs[i].PubKeys), nil
}

func (st *TxStore) indexTxos(txs []*bcpb.Tx) error {
	for _, tx := range txs {
		for i, txo := range tx.Outputs {
			err := st.dki.Set(txo.DataKey, tx.Digest, int32(i))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
