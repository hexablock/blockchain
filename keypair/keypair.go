package keypair

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/hasher"
)

//const version byte = 0x00
const addressChecksumLen = 4

// KeyPair holds a public private keypair
type KeyPair struct {
	h     hasher.Hasher
	curve elliptic.Curve
	// Raw private and public key
	PrivateKey ecdsa.PrivateKey
	// Public key bytes for the private key.
	PublicKey bcpb.PublicKey
}

// New returns a new empty keypair populated with the curve and hasher
func New(curve elliptic.Curve, h hasher.Hasher) *KeyPair {
	return &KeyPair{
		h:     h,
		curve: curve,
	}
}

// Generate creates and returns a KeyPair
func Generate(curve elliptic.Curve, h hasher.Hasher) (*KeyPair, error) {
	kp := New(curve, h)
	err := kp.generate()
	return kp, err
}

func (w *KeyPair) generate() error {
	private, err := ecdsa.GenerateKey(w.curve, rand.Reader)
	if err == nil {
		w.PrivateKey = *private
		w.setPublicKey()
	}
	return err
}

func (w *KeyPair) setPublicKey() {
	priv := w.PrivateKey
	pubkey := append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
	w.PublicKey = bcpb.PublicKey(pubkey)
}

// Algorithm returns the keypair algorithm
func (w KeyPair) Algorithm() []byte {
	return []byte(fmt.Sprintf("ecdsa%d", w.curve.Params().BitSize))
}

// Address returns the public key address
func (w KeyPair) Address() []byte {
	return w.PublicKey.Address(w.h)
}

// Sign signs the digest and returns the signature
func (w KeyPair) Sign(digest bcpb.Digest) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, &w.PrivateKey, digest)
	if err == nil {
		return append(r.Bytes(), s.Bytes()...), nil
	}

	return nil, err
}

// VerifySignature verifies the signature for te digest
func (w KeyPair) VerifySignature(digest bcpb.Digest, signature []byte) bool {

	r := big.Int{}
	s := big.Int{}
	sigLen := len(signature)
	r.SetBytes(signature[:(sigLen / 2)])
	s.SetBytes(signature[(sigLen / 2):])

	pubkey := w.PublicKey

	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubkey)
	x.SetBytes(pubkey[:(keyLen / 2)])
	y.SetBytes(pubkey[(keyLen / 2):])

	rawPubKey := ecdsa.PublicKey{Curve: w.curve, X: &x, Y: &y}

	return ecdsa.Verify(&rawPubKey, digest, &r, &s)
}

// Save x509 marshals the key and writes it to the given path
func (w KeyPair) Save(fpath string) error {
	data, err := x509.MarshalECPrivateKey(&w.PrivateKey)
	if err == nil {
		err = ioutil.WriteFile(fpath, data, 0644)
	}

	return err
}

// FromFile loads an existing keypair from the given filepath
func FromFile(fpath string) (*KeyPair, error) {
	der, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	var kp *KeyPair
	key, err := x509.ParseECPrivateKey(der)
	if err == nil {
		kp = &KeyPair{
			PrivateKey: *key,
			curve:      key.Curve,
			h:          hasher.Default(),
		}
		kp.setPublicKey()
	}

	return kp, err
}
