package stores

import (
	"bytes"
	"errors"

	"github.com/dgraph-io/badger"
	proto "github.com/gogo/protobuf/proto"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/hasher"
)

const (
	// Over all key prefix
	blkSubkeyPrefix = "blk/"
	// Genesis block key sub prefix appended to blkSubkeyPrefix
	blkGenesisSubkeyPrefix = "genesis"
	// Last block key sub prefix appended to blkSubkeyPrefix
	blkLastSubkeyPrefix = "last"
	// Last executed block key sub prefix appended to blkSubkeyPrefix
	blkExecSubkeyPrefix = "exec"
)

var (
	// ErrBlockExists is used when writing a block that already exists
	ErrBlockExists   = errors.New("block exists")
	ErrBlockNotFound = errors.New("block not found")
)

// BlockIterator is used to iterate over blocks in a store
type BlockIterator func(bcpb.Digest, *bcpb.Block) error

type BadgerBlockStorage struct {
	hasher hasher.Hasher

	// prefix for all stored keys
	prefix []byte

	db *badger.DB
}

func NewBadgerBlockStorage(db *badger.DB, keyPrefix []byte, h hasher.Hasher) *BadgerBlockStorage {
	return &BadgerBlockStorage{
		db:     db,
		hasher: h,
		prefix: append(keyPrefix, []byte(blkSubkeyPrefix)...),
	}

}

func (s *BadgerBlockStorage) Close() error {
	return s.db.Close()
}

func (st *BadgerBlockStorage) Get(id bcpb.Digest) (*bcpb.Block, error) {
	key := st.getkey(id)

	var data []byte
	err := st.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(nil)
		return err
	})

	var blk bcpb.Block
	if err == nil {
		err = proto.Unmarshal(data, &blk)
	}
	return &blk, err
}

func (st *BadgerBlockStorage) Genesis() (id bcpb.Digest, blk *bcpb.Block) {
	st.db.View(func(txn *badger.Txn) error {
		var err error
		id, blk, err = st.getGenesisBlock(txn)
		return err
	})

	return id, blk
}

func (st *BadgerBlockStorage) Last() (id bcpb.Digest, blk *bcpb.Block) {
	st.db.View(func(txn *badger.Txn) error {
		var err error
		id, blk, err = st.getLastBlock(txn)
		return err
	})

	return id, blk
}

func (st *BadgerBlockStorage) LastExec() (id bcpb.Digest, blk *bcpb.Block) {
	st.db.View(func(txn *badger.Txn) error {
		var err error
		id, blk, err = st.getPointerBlock([]byte(blkExecSubkeyPrefix), txn)
		return err
	})

	return id, blk
}

func (st *BadgerBlockStorage) Exists(id bcpb.Digest) bool {
	key := st.getkey(id)
	err := st.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		return err
	})
	return (err == nil)
}

func (st *BadgerBlockStorage) Add(b *bcpb.Block) (bcpb.Digest, error) {
	id := b.Header.Hash(st.hasher.Clone())
	key := st.getkey(id)

	buf, err := proto.Marshal(b)
	if err != nil {
		return nil, err
	}

	err = st.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return ErrBlockExists
		}

		return txn.Set(key, buf)
	})

	return id, err
}

func (st *BadgerBlockStorage) SetGenesis(id bcpb.Digest) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Set(st.getkey([]byte(blkGenesisSubkeyPrefix)), id)
	})
}

func (st *BadgerBlockStorage) SetLast(id bcpb.Digest) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Set(st.getkey([]byte(blkLastSubkeyPrefix)), id)
	})
}

func (st *BadgerBlockStorage) SetLastExec(id bcpb.Digest) error {
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Set(st.getkey([]byte(blkExecSubkeyPrefix)), id)
	})
}

func (st *BadgerBlockStorage) Remove(id bcpb.Digest) error {
	key := st.getkey(id)
	return st.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (st *BadgerBlockStorage) Iter(f BlockIterator) error {
	prefix := st.getkey([]byte(st.hasher.Name()))

	return st.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		iter := txn.NewIterator(opt)

		for iter.Seek(prefix); iter.Valid(); iter.Next() {
			item := iter.Item()
			key := item.Key()
			if !bytes.HasPrefix(key, prefix) {
				break
			}

			value, err := item.Value()
			if err != nil {
				return err
			}

			bid := bytes.TrimPrefix(item.Key(), st.prefix)

			var block bcpb.Block
			if err = proto.Unmarshal(value, &block); err != nil {
				return err
			}

			if err = f(bid, &block); err != nil {
				return err
			}
		}

		return nil
	})
}

func (st *BadgerBlockStorage) getkey(key []byte) []byte {
	return append(st.prefix, key...)
}

func (st *BadgerBlockStorage) getGenesisBlock(txn *badger.Txn) (bcpb.Digest, *bcpb.Block, error) {
	return st.getPointerBlock([]byte(blkGenesisSubkeyPrefix), txn)
}

func (st *BadgerBlockStorage) getLastBlock(txn *badger.Txn) (bcpb.Digest, *bcpb.Block, error) {
	return st.getPointerBlock([]byte(blkLastSubkeyPrefix), txn)
}

func (st *BadgerBlockStorage) getPointerBlock(id []byte, txn *badger.Txn) (bcpb.Digest, *bcpb.Block, error) {
	lastBlockKey := st.getkey(id)

	lItm, err := txn.Get(lastBlockKey)
	if err != nil {
		return nil, nil, err
	}
	lbId, err := lItm.Value()
	if err != nil {
		return nil, nil, err
	}

	lid := bcpb.Digest(lbId)

	lbIdKey := st.getkey(lbId)
	lbItm, err := txn.Get(lbIdKey)
	if err != nil {
		return lid, nil, err
	}
	lbVal, err := lbItm.ValueCopy(nil)
	if err != nil {
		return lid, nil, err
	}

	var blk bcpb.Block
	err = proto.Unmarshal(lbVal, &blk)
	return lid, &blk, err
}

func cloneBlk(blk *bcpb.Block) *bcpb.Block {
	b := *blk
	hdr := *blk.Header
	b.Header = &hdr
	return &b
}
