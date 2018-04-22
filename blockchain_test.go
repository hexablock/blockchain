package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/keypair"
	"github.com/hexablock/blockchain/stores"
)

func testBlockchainConf() *Config {
	conf := DefaultConfig()

	conf.BlockStorage = stores.NewBadgerBlockStorage(testDB, []byte("test-prefix/"), conf.Hasher)
	conf.TxStorage = stores.NewBadgerTxStorage(testDB, []byte("test-prefix/"))
	conf.DataKeyIndex = stores.NewBadgerDataKeyIndex(testDB, []byte("test-prefix/"))

	return conf
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
	genesis := NewGenesisBlock(conf.Hasher)
	genesis.SetTxs(txs, bc.Hasher())

	// Check error
	err := genesis.Sign(kp.PublicKey, []byte("foo"))
	assert.Equal(t, bcpb.ErrSignerNotInBlock, err)

	genesis.SetProposer(kp.PublicKey)
	genesis.SetHash(bc.Hasher())

	// Check sign
	signature, err := kp.Sign(genesis.Digest)
	assert.Nil(t, err)

	err = genesis.Sign(kp.PublicKey, signature)
	assert.Nil(t, err)

	err = bc.SetGenesis(genesis, txs)
	assert.Nil(t, err)
	err = bc.Commit(genesis.Digest)
	assert.Nil(t, err)
	err = bc.SetLastExec(genesis.Digest)
	assert.Nil(t, err)

	g := bc.Genesis()
	assert.Equal(t, genesis.Digest, g.Digest)

	var c int
	bc.tx.dki.Iter(nil, func(dk bcpb.DataKey, ref bcpb.Digest, i int32) bool {
		c++
		return true
	})
	assert.Equal(t, 1, c)

	tx1 := bcpb.NewTx()

	txref, _, err := bc.tx.GetDataKeyTx(bcpb.DataKey("test:key"))
	assert.Nil(t, err)

	txi1, err := bc.NewTxInput(bcpb.DataKey("test:key"))
	assert.Nil(t, err)
	//txi1 := bcpb.NewTxInput(txref.Digest, i, nil)
	tx1.AddInput(txi1)
	tx1.AddOutput(&bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key"),
	})
	tx1.SetDigest(bc.Hasher())

	tx2 := bcpb.NewBaseTx()
	tx2.AddOutput(&bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key1"),
	})
	tx2.SetDigest(bc.Hasher())

	nb := nextBlock(bc.blk)
	txs1 := []*bcpb.Tx{tx1, tx2}
	nb.SetTxs(txs1, conf.Hasher)
	nb.SetHash(bc.Hasher())

	lid, err := bc.Append(nb, txs1)
	assert.Nil(t, err)

	err = bc.Commit(lid)
	assert.Nil(t, err)

	txref, i, err := bc.tx.GetDataKeyTx(bcpb.DataKey("test:key"))
	assert.Nil(t, err)
	assert.Equal(t, int32(0), i)
	assert.Equal(t, tx1.Digest, txref.Digest)

	last := bc.Last()
	assert.Equal(t, lid, last.Digest)

	// Sigs
	tx3 := bcpb.NewTx()

	txi3, err := bc.NewTxInput(bcpb.DataKey("test:key"))
	txi3.AddPubKey(kp.PublicKey)

	assert.Nil(t, err)
	assert.False(t, txi3.IsBase())

	digest := txi3.Hash(bc.Hasher())
	sig, _ := kp.Sign(digest)
	txi3.Sign(kp.PublicKey, sig)

	tx3.AddInput(txi3)

	txo3 := &bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key"),
		PubKeys: []bcpb.PublicKey{kp.PublicKey},
	}
	txo3.SetRequiredSignatures(1)
	tx3.AddOutput(txo3)
	tx3.SetDigest(bc.Hasher())

	nb = nextBlock(bc.blk)
	txs3 := []*bcpb.Tx{tx3}
	nb.SetTxs(txs3, conf.Hasher)
	nb.SetHash(bc.Hasher())

	lid, err = bc.Append(nb, txs3)
	assert.Nil(t, err)

	err = bc.Commit(lid)
	assert.Nil(t, err)

	//
	// Check logic
	//
	tx4 := bcpb.NewTx()

	txi4, _ := bc.NewTxInput(bcpb.DataKey("test:key"))
	txi4.AddPubKey(kp.PublicKey)

	digest = txi4.Hash(bc.Hasher())
	sig, _ = kp.Sign(digest)
	txi4.Sign(kp.PublicKey, sig)

	tx4.AddInput(txi4)

	txo4 := &bcpb.TxOutput{
		DataKey: bcpb.DataKey("test:key"),
		PubKeys: []bcpb.PublicKey{kp.PublicKey},
	}
	txo4.SetRequiredSignatures(1)
	tx4.AddOutput(txo4)
	tx4.SetDigest(bc.Hasher())

	nb = nextBlock(bc.blk)
	txs4 := []*bcpb.Tx{tx4}
	nb.SetTxs(txs4, conf.Hasher)
	nb.SetHash(bc.Hasher())

	lid, err = bc.Append(nb, txs4)
	assert.Nil(t, err)

	err = bc.Commit(lid)
	assert.Nil(t, err)

}
