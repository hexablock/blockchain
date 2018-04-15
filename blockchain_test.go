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
}
