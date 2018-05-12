package stores

import (
	"bytes"
	"encoding/binary"

	"github.com/dgraph-io/badger"
	"github.com/hexablock/blockchain/bcpb"
)

const (
	idxSubkeyPrefix = "idx/"
)

// DataKeyIterator is used to iterate over the datakey index
type DataKeyIterator func(bcpb.DataKey, bcpb.Digest, int32) bool

// BadgerDataKeyIndex implements the DataKeyIndex interface backed by badger
// key-value store
type BadgerDataKeyIndex struct {
	db     *badger.DB
	prefix []byte
}

// NewBadgerDataKeyIndex inits a new BadgerDataKeyIndex
func NewBadgerDataKeyIndex(db *badger.DB, prefix []byte) *BadgerDataKeyIndex {
	return &BadgerDataKeyIndex{
		db:     db,
		prefix: append(prefix, []byte(idxSubkeyPrefix)...),
	}
}

func (index *BadgerDataKeyIndex) getkey(key bcpb.DataKey) []byte {
	return append(index.prefix, key...)
}

// Get retrieves the digest and output index associated to the DataKey
func (index *BadgerDataKeyIndex) Get(key bcpb.DataKey) (bcpb.Digest, int32, error) {
	var (
		i      int32
		digest bcpb.Digest
		k      = index.getkey(key)
	)

	err := index.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		i = int32(binary.BigEndian.Uint32(val[:4]))
		digest = bcpb.Digest(val[4:])
		return nil
	})

	return digest, i, err
}

// Set sets the DataKey to the given tx id and output index
func (index *BadgerDataKeyIndex) Set(key bcpb.DataKey, ref bcpb.Digest, idx int32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(idx))
	k := index.getkey(key)

	return index.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, append(b, ref...))
	})
}

// Iter iterates over all DataKeys starting at the given prefix DataKey
func (index *BadgerDataKeyIndex) Iter(prefix bcpb.DataKey, f DataKeyIterator) error {
	pfx := index.getkey(prefix)

	return index.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)

		for iter.Seek(pfx); iter.Valid(); iter.Next() {
			item := iter.Item()
			key := item.Key()
			if !bytes.HasPrefix(key, pfx) {
				break
			}
			k := bytes.TrimPrefix(key, index.prefix)

			val, err := item.Value()
			if err != nil {
				return err
			}

			i := int32(binary.BigEndian.Uint32(val[:4]))
			digest := bcpb.Digest(val[4:])

			if !f(bcpb.DataKey(k), digest, i) {
				break
			}
		}

		return nil
	})
}
