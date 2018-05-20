package stores

import (
	"bytes"
	"errors"
	"log"

	"github.com/dgraph-io/badger"
	proto "github.com/gogo/protobuf/proto"

	"github.com/hexablock/blockchain/bcpb"
)

const (
	// Transaction key sub prefix
	txSubkeyPrefix = "tx/"
)

var (
	errTxExists   = errors.New("tx exists")
	errTxNotFound = errors.New("tx not found")
)

// BadgerTxStorage implements a badger backed TxStorage interface
type BadgerTxStorage struct {
	prefix []byte
	db     *badger.DB
}

// NewBadgerTxStorage returns a new badger backed tx storage device.
func NewBadgerTxStorage(db *badger.DB, keyPrefix []byte) *BadgerTxStorage {
	return &BadgerTxStorage{
		prefix: append(keyPrefix, []byte(txSubkeyPrefix)...),
		db:     db,
	}
}

func (store *BadgerTxStorage) getkey(key []byte) []byte {
	return append(store.prefix, key...)
}

// Get retrieves a transaction by the given id
func (store *BadgerTxStorage) Get(id bcpb.Digest) (*bcpb.Tx, error) {
	var (
		key = store.getkey(id)
		val []byte
	)

	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	var tx bcpb.Tx
	err = proto.Unmarshal(val, &tx)
	return &tx, err

}

// Iter iterates over each transaction, calling the f for each encountered
// tx.
func (store *BadgerTxStorage) Iter(f func(bcpb.Tx) error) {
	store.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)

		for iter.Seek(store.prefix); iter.Valid(); iter.Next() {

			item := iter.Item()
			key := item.Key()

			if !bytes.HasPrefix(key, store.prefix) {
				break
			}

			val, err := item.Value()
			if err != nil {
				log.Printf("[ERR] Failed to get value key=%q: %v", key, err)
				continue
			}

			var tx bcpb.Tx
			err = proto.Unmarshal(val, &tx)
			if err != nil {
				log.Printf("[ERR] Failed to unmarshal tx key=%q: %v", key, err)
				continue
			}

			if err := f(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// Set sets the transaction returning the hash id of the transaction
func (store *BadgerTxStorage) Set(tx *bcpb.Tx) error {
	b, err := proto.Marshal(tx)
	if err != nil {
		return err
	}

	key := store.getkey(tx.Digest)
	err = store.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, b)
	})

	return err
}

// SetBatch sets a batch of transactions returning each id or an error
func (store *BadgerTxStorage) SetBatch(txs []*bcpb.Tx) error {
	var (
		l      = len(txs)
		keys   = make([]bcpb.Digest, l)
		values = make([][]byte, l)
		err    error
	)

	for i := range txs {
		values[i], err = proto.Marshal(txs[i])
		if err != nil {
			return err
		}

		keys[i] = store.getkey(txs[i].Digest)

	}

	err = store.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < l; i++ {
			if er := txn.Set(keys[i], values[i]); er != nil {
				return er
			}
		}
		return nil
	})

	return err
}
