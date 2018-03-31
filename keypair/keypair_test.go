package keypair

import (
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/base58"
	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/hasher"
)

func Test_KeyPair(t *testing.T) {
	kp, _ := Generate(elliptic.P256(), hasher.Default())
	assert.Equal(t, "ecdsa256", string(kp.Algorithm()))

	b58 := base58.Encode(kp.PublicKey)

	pubkey := base58.Decode(b58)
	assert.Equal(t, kp.PublicKey, bcpb.PublicKey(pubkey))

}
