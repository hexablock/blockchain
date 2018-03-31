package keypair

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/hexablock/blockchain/bcpb"
	"github.com/hexablock/blockchain/hasher"
)

//const version byte = 0x00
const addressChecksumLen = 4

type KeyPair struct {
	h          hasher.Hasher
	curve      elliptic.Curve
	PrivateKey ecdsa.PrivateKey
	PublicKey  bcpb.PublicKey
}

func New(curve elliptic.Curve, h hasher.Hasher) *KeyPair {
	return &KeyPair{
		h:     h,
		curve: curve,
	}
}

// Generate creates and returns a KeyPair
func Generate(curve elliptic.Curve, h hasher.Hasher) (*KeyPair, error) {
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	kp := &KeyPair{
		h,
		curve,
		*private,
		bcpb.PublicKey(pubKey),
	}

	return kp, nil
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

// // Address returns KeyPair address.
// func (w KeyPair) Address() []byte {
// 	pubKeyHash := w.pubkeyHashRipeMD()
//
// 	versionedPayload := append([]byte{version}, pubKeyHash...)
// 	checksum := checksum(versionedPayload, w.h.New())
//
// 	fullPayload := append(versionedPayload, checksum...)
// 	address := base58.Encode(fullPayload)
//
// 	return address
// }

// func (w KeyPair) pubkeyHashRipeMD() []byte {
// 	// Hash public key with supplied hash function
// 	h := w.h.New()
// 	h.Write(w.PublicKey)
// 	sh := h.Sum(nil)
//
// 	// Ripemd the hash
// 	RIPEMD160Hasher := ripemd160.New()
// 	RIPEMD160Hasher.Write(sh[:])
// 	return RIPEMD160Hasher.Sum(nil)
// }
//
// func ValidateAddress(address string, h hash.Hash) bool {
// 	pubKeyHash := base58.Decode([]byte(address))
// 	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
// 	version := pubKeyHash[0]
// 	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
// 	targetChecksum := checksum(append([]byte{version}, pubKeyHash...), h)
//
// 	return bytes.Compare(actualChecksum, targetChecksum) == 0
// }

// // Checksum generates a checksum for a public key
// func checksum(payload []byte, h hash.Hash) []byte {
// 	// First hash
// 	h.Write(payload)
// 	sh1 := h.Sum(nil)
//
// 	// 2nd hash i.e. Hash of hash
// 	h.Reset()
// 	h.Write(sh1)
// 	sh2 := h.Sum(nil)
//
// 	return sh2[:addressChecksumLen]
// }
