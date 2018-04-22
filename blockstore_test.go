package blockchain

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/stores"
	"github.com/hexablock/hasher"
)

var testDB *badger.DB

func testBadgerDB(tmpdir string) (*badger.DB, error) {
	opt := badger.DefaultOptions
	opt.Dir = tmpdir
	opt.ValueDir = tmpdir
	return badger.Open(opt)
}

func TestMain(m *testing.M) {
	tdir, _ := ioutil.TempDir("/tmp", "blockstore")
	defer os.RemoveAll(tdir)

	db, err := testBadgerDB(tdir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	testDB = db

	code := m.Run()

	os.Exit(code)
}

func nextBlock(st *blockStore) *bcpb.Block {
	lid, last := st.st.Last()

	blk := bcpb.NewBlock()
	blk.Header = &bcpb.BlockHeader{
		Height:    last.Header.Height + 1,
		PrevBlock: lid,
		Timestamp: time.Now().UnixNano(),
		Nonce:     last.Header.Nonce + 1,
	}

	return blk
}

func Test_BlockStore(t *testing.T) {
	h := hasher.Default()
	bdb := stores.NewBadgerBlockStorage(testDB, []byte("test/"), h)

	bs := &blockStore{bdb}

	// Genesis
	btx := bcpb.NewBaseTx()
	genesis := NewGenesisBlock(h)
	genesis.SetTxs([]*bcpb.Tx{btx}, h)
	genesis.SetHash(h)

	err := bs.SetGenesis(genesis)
	assert.Nil(t, err)
	bs.st.SetLast(genesis.Digest)
	bs.st.SetLastExec(genesis.Digest)

	lid1, last := bdb.Last()
	assert.Equal(t, uint32(0), last.Height())

	err = bs.SetGenesis(genesis)
	assert.NotNil(t, err)

	// First block
	blk := nextBlock(bs)
	blk.SetTxs([]*bcpb.Tx{bcpb.NewBaseTx()}, h)

	id, err := bs.Append(blk)
	assert.Nil(t, err)
	err = bs.st.SetLast(id)
	assert.Nil(t, err)

	lid2, last := bdb.Last()
	assert.Equal(t, uint32(1), last.Height())
	assert.Equal(t, uint64(2), last.Header.Nonce)

	// Check errors
	b2 := *blk
	b2.Header.Height = last.Height() + 1
	b2.Header.PrevBlock = lid2
	b2.Header.Nonce = 1
	_, err = bs.Append(&b2)
	assert.Equal(t, errInvalidNonce, err)

	b2.Header.Height = 0
	_, err = bs.Append(&b2)
	assert.Equal(t, errHeightMismatch, err)

	b3 := nextBlock(bs)
	b3.Header.PrevBlock = lid1
	_, err = bs.Append(b3)
	assert.Equal(t, ErrPrevBlockMismatch, err)

}
