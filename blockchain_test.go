package blockchain

import (
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/hasher"
	"github.com/hexablock/blockchain/keypair"
	"github.com/hexablock/blockchain/stores"
)

func testBlockchainConf() *Config {
	h := hasher.Default()
	return &Config{
		h,
		elliptic.P256(),
		nil,
		stores.NewBadgerBlockStorage(testDB, []byte("test-prefix/"), h),
		stores.NewBadgerTxStorage(testDB, []byte("test-prefix/")),
		stores.NewBadgerDataKeyIndex(testDB, []byte("test-prefix/")),
	}
}

func Test_Blockchain(t *testing.T) {
	conf := testBlockchainConf()

	kp, _ := keypair.Generate(conf.Curve, conf.Hasher)
	bc := New(conf)

	tx := bcpb.NewBaseTx()
	txo := &bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key"),
	}
	tx.AddOutput(txo)

	txs := []*bcpb.Tx{tx}
	genesis := NewGenesisBlock(txs, conf.Hasher)

	// Check error
	err := genesis.Sign(kp.PublicKey, []byte("foo"))
	assert.Equal(t, bcpb.ErrSignerNotInBlock, err)

	// Check sign
	genesis.AddSigner(kp.PublicKey)
	digest := genesis.Header.Hash(conf.Hasher)
	signature, err := kp.Sign(digest)
	assert.Nil(t, err)

	err = genesis.Sign(kp.PublicKey, signature)
	assert.Nil(t, err)

	err = bc.SetGenesis(genesis, txs)
	assert.Nil(t, err)

	var c int
	bc.tx.dki.Iter(nil, func(dk bcpb.DataKey, ref bcpb.Digest, i int32) bool {
		c++
		return true
	})
	assert.Equal(t, 1, c)

	tx1 := bcpb.NewTx()

	txref, i, err := bc.tx.GetDataKeyTx(bcpb.DataKey("test:key"))
	assert.Nil(t, err)

	txi1 := bcpb.NewTxInput(txref.Digest, i, nil)
	tx1.AddInput(txi1)
	tx1.AddOutput(&bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key"),
	})
	tx1.SetDigest(conf.Hasher)

	tx2 := bcpb.NewBaseTx()
	tx2.AddOutput(&bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key1"),
	})
	tx2.SetDigest(conf.Hasher)

	nb := bc.blk.NextBlock()
	txs1 := []*bcpb.Tx{tx1, tx2}
	nb.SetTxs(txs1, conf.Hasher)

	lid, err := bc.Append(nb, txs1)
	assert.Nil(t, err)

	err = bc.Commit(lid)
	assert.Nil(t, err)

	txref, i, err = bc.tx.GetDataKeyTx(bcpb.DataKey("test:key"))
	assert.Nil(t, err)
	assert.Equal(t, int32(0), i)
	assert.Equal(t, tx1.Digest, txref.Digest)

}
