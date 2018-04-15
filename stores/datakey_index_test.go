package stores

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/hasher"
)

func Test_DataKeyIndex(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("/tmp", "ora-tx-store-")
	defer os.RemoveAll(tmpdir)

	db, err := testBadgerDB(tmpdir)
	assert.Nil(t, err)
	defer db.Close()

	idx := NewBadgerDataKeyIndex(db, []byte("foo"))

	key := bcpb.DataKey("/ball")
	z := bcpb.NewZeroDigest(hasher.Default())
	err = idx.Set(key, z, 0)
	assert.Nil(t, err)

	ref, i, err := idx.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, int32(0), i)
	assert.Equal(t, z, ref)

	for i := 0; i < 5; i++ {
		k := bcpb.DataKey(fmt.Sprintf("/nums/%d", i))
		z := bcpb.NewZeroDigest(hasher.Default())
		err = idx.Set(k, z, int32(i))
		assert.Nil(t, err)
	}

	var c int
	idx.Iter(bcpb.DataKey(""), func(k bcpb.DataKey, ref bcpb.Digest, i int32) bool {
		c++
		return true
	})
	assert.Equal(t, 6, c)

	c = 0
	idx.Iter(bcpb.DataKey("/nums"), func(k bcpb.DataKey, ref bcpb.Digest, i int32) bool {
		c++
		return true
	})
	assert.Equal(t, 5, c)
}

func testBadgerDB(tmpdir string) (*badger.DB, error) {
	opt := badger.DefaultOptions
	opt.Dir = tmpdir
	opt.ValueDir = tmpdir
	return badger.Open(opt)
}
