package blockchain

import (
	"crypto/elliptic"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/keypair"
	"github.com/hexablock/blockchain/stores"
	"github.com/hexablock/hasher"
)

func Test_TxStore(t *testing.T) {
	h := hasher.Default()
	bst := stores.NewBadgerTxStorage(testDB, []byte("test/"))
	ist := stores.NewBadgerDataKeyIndex(testDB, []byte("idx/"))
	st := &txStore{bst, ist}

	btx := bcpb.NewBaseTx()
	btx.SetDigest(h)

	err := st.tx.Set(btx)
	assert.Nil(t, err)

	_, err = st.tx.Get(btx.Digest)
	assert.Nil(t, err)

	fkey := bcpb.DataKey("foo:bar")
	_, _, err = st.GetDataKeyTx(fkey)
	assert.NotNil(t, err)

	_, err = st.NewTxInput(fkey)
	assert.NotNil(t, err)
}

func Test_TxStore_Find(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("/tmp", "txfinder-")
	defer os.RemoveAll(tmpdir)

	db, err := testBadgerDB(tmpdir)
	assert.Nil(t, err)
	defer db.Close()

	bst := stores.NewBadgerTxStorage(db, []byte("txfinder/"))
	txstore := &txStore{bst, nil}

	h := hasher.Default()
	kp1, _ := keypair.Generate(elliptic.P256(), h)
	kp2, _ := keypair.Generate(elliptic.P256(), h)

	// fmt.Println("1", string(PublicKey(kp1.PublicKey).Address(conf.Hasher)))
	// fmt.Println("2", string(PublicKey(kp2.PublicKey).Address(conf.Hasher)))
	// fmt.Println()

	//txids := make([]bcpb.Digest, 12)

	// First tx
	txn := bcpb.NewTx()
	txn.AddInput(bcpb.NewBaseTxInput([]byte("first-tx")))
	txo := &bcpb.TxOutput{Data: []byte("first")}
	txo.PubKeys = []bcpb.PublicKey{kp1.PublicKey, kp2.PublicKey}
	txn.AddOutput(txo)

	txn.SetDigest(h)
	err = txstore.tx.Set(txn)
	assert.Nil(t, err)
	//fmt.Println("ADDED", txids[0].String())

	gtx, err := txstore.tx.Get(txn.Digest)
	assert.Nil(t, err)
	assert.NotNil(t, gtx)

	// Second tx
	tx1 := bcpb.NewTx()
	txi := &bcpb.TxInput{Ref: txn.Digest, Index: 0,
		PubKeys: []bcpb.PublicKey{kp2.PublicKey}}
	tx1.AddInput(txi)

	txo1 := &bcpb.TxOutput{Data: []byte("2nd")}
	txo1.PubKeys = []bcpb.PublicKey{kp1.PublicKey, kp2.PublicKey}
	tx1.AddOutput(txo1)

	tx1.SetDigest(h)
	err = txstore.tx.Set(tx1)
	assert.Nil(t, err)
	//fmt.Println("ADDED", txids[1].String())

	gtx, err = txstore.tx.Get(tx1.Digest)
	assert.Nil(t, err)
	assert.NotNil(t, gtx)

	var c int
	txstore.tx.Iter(func(arg2 bcpb.Tx) error {
		c++
		return nil
	})
	assert.Equal(t, 2, c)

	txids := make([]bcpb.Digest, 6)
	txids[0] = txn.Digest
	txids[1] = tx1.Digest
	j := 2
	for i := 0; i < 4; i++ {
		txn := bcpb.NewTx()
		txo := &bcpb.TxOutput{Data: []byte{byte(i)}}

		ref := txids[j-1]
		//fmt.Println("REFD", txids[j-1].String())
		txi := &bcpb.TxInput{Ref: ref, Index: 0}

		if i%2 == 0 {
			txi.PubKeys = []bcpb.PublicKey{kp1.PublicKey}
			txo.PubKeys = []bcpb.PublicKey{kp2.PublicKey}
			//fmt.Printf("%d to %x\n", i, txo.PubKeys[0])
		} else {
			txi.PubKeys = []bcpb.PublicKey{kp2.PublicKey}
			txo.PubKeys = []bcpb.PublicKey{kp1.PublicKey}
			//fmt.Printf("%d to %x\n", i, txo.PubKeys[0])
		}

		txn.AddInput(txi)
		txn.AddOutput(txo)
		txn.SetDigest(h)
		txids[j] = txn.Digest.Copy()
		txstore.tx.Set(txn)

		j++
	}

	unspent := txstore.FindUnspent()
	assert.Equal(t, 1, len(unspent))

	var txn1 bcpb.Tx
	for k := range unspent {
		txn1 = unspent[k]
		break
	}

	usableOuts := txstore.FindUTX(kp1.PublicKey)
	assert.Equal(t, 1, len(usableOuts))
	var txn2 bcpb.Tx
	for k := range usableOuts {
		txn2 = usableOuts[k]
		break
	}

	assert.Equal(t, txn1.Digest, txn2.Digest)

}
