package keypair

import (
	"crypto/elliptic"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/base58"
	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/hasher"
)

func Test_KeyPair(t *testing.T) {
	kp, _ := Generate(elliptic.P256(), hasher.Default())
	assert.Equal(t, "ecdsa256", string(kp.Algorithm()))

	b58 := base58.Encode(kp.PublicKey)

	pubkey := base58.Decode(b58)
	pk := bcpb.PublicKey(pubkey)
	assert.Equal(t, kp.PublicKey, pk)
	assert.Equal(t, kp.Address(), pk.Address(hasher.Default()))

	sig, err := kp.Sign(bcpb.Digest("xxxx"))
	assert.Nil(t, err)
	v := kp.VerifySignature(bcpb.Digest("xxxx"), sig)
	assert.True(t, v)

	tmpfile, _ := ioutil.TempFile("/tmp", "kptest-")
	tmpfile.Close()
	err = kp.Save(tmpfile.Name())
	assert.Nil(t, err)

	_, err = FromFile(tmpfile.Name())
	assert.Nil(t, err)

}

func Test_FromFile_error(t *testing.T) {
	_, err := FromFile("foo/barbadfd/dfdf")
	assert.NotNil(t, err)
}
