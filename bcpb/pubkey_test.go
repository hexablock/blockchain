package bcpb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexablock/hasher"
)

func Test_PublicKey(t *testing.T) {
	pk := PublicKey("qazxswedcvfrtgbnhyujmkiolop")
	h := hasher.Default()
	addr := pk.Address(h)

	valid := ValidatePublicKeyAddress(addr, h.New())
	assert.True(t, valid)
}
