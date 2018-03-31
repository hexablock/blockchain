package blockchain

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/hasher"
	"github.com/hexablock/blockchain/stores"
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

func Test_BlockStore(t *testing.T) {
	h := hasher.Default()
	bdb := stores.NewBadgerBlockStorage(testDB, []byte("test/"), h)

	bs := &BlockStore{bdb}

	// Genesis
	btx := bcpb.NewBaseTx()
	genesis := NewGenesisBlock([]*bcpb.Tx{btx}, h)

	err := bs.SetGenesis(genesis)
	assert.Nil(t, err)

	lid1, last := bdb.Last()
	assert.Equal(t, uint32(0), last.Height())

	err = bs.SetGenesis(genesis)
	assert.NotNil(t, err)

	// First block
	blk := bs.NextBlock()
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

	b2.Header.PrevBlock = lid1
	_, err = bs.Append(&b2)
	assert.Equal(t, errPrevBlockMismatch, err)

	b2.Header.Height = 0
	_, err = bs.Append(&b2)
	assert.Equal(t, errHeightMismatch, err)
}
