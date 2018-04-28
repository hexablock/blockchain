package bcpb

import (
	"testing"

	"github.com/hexablock/hasher"
	"github.com/stretchr/testify/assert"
)

func Test_TxInput(t *testing.T) {

	txi := NewTxInput(nil, -1, []PublicKey{PublicKey("foo")})
	txi.AddArgs([]byte("bar"))
	assert.Equal(t, 2, len(txi.Signatures))
	assert.Equal(t, 1, len(txi.Args()))
	assert.True(t, txi.IsBase())

	txi.AddPubKey(PublicKey("baz"))
	assert.Equal(t, 3, len(txi.Signatures))

	assert.Nil(t, txi.Signatures[0])
	assert.Nil(t, txi.Signatures[1])
	assert.Equal(t, []byte("bar"), txi.Signatures[2])

	assert.False(t, txi.AddPubKey(PublicKey("foo")))
	assert.Nil(t, txi.Sign(PublicKey("foo"), []byte("sig")))
	assert.Equal(t, ErrNotAuthorized, txi.Sign(PublicKey("notfound"), []byte("sig")))

	digest := txi.Hash(hasher.Default())
	_, err := ParseDigest(digest.String())
	assert.Nil(t, err)

	btxi := NewBaseTxInput()
	assert.Equal(t, 0, len(btxi.Signatures))
}
