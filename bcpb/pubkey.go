package bcpb

import (
	"bytes"
	"hash"

	"golang.org/x/crypto/ripemd160"

	"github.com/hexablock/blockchain/base58"
	"github.com/hexablock/blockchain/hasher"
)

const addressChecksumLen = 4

// PublicKey contains public key bytes
type PublicKey []byte

// Equal returns true if both keys are the same
func (w PublicKey) Equal(pk PublicKey) bool {
	return bytes.Compare(w, pk) == 0
}

// Address returns the hashed key address
func (w PublicKey) Address(h hasher.Hasher) []byte {
	pubKeyHash := w.pubkeyHashRipeMD(h)

	checksum := checksum(pubKeyHash, h.New())
	fullPayload := append(pubKeyHash, checksum...)

	return base58.Encode(fullPayload)
}

// Hash generates the hash of the public key
func (w PublicKey) Hash(hf hasher.Hasher) Digest {
	h := hf.New()
	h.Write(w)
	sh := h.Sum(nil)
	return NewDigest(hf.Name(), sh)
}

func (w PublicKey) pubkeyHashRipeMD(hf hasher.Hasher) []byte {
	// Hash public key with supplied hash function
	digest := w.Hash(hf)

	// Ripemd the hash
	rmd := ripemd160.New()
	rmd.Write(digest[:])
	return rmd.Sum(nil)
}

// Checksum generates a checksum for a public key
func checksum(payload []byte, h hash.Hash) []byte {
	// First hash
	h.Write(payload)
	sh1 := h.Sum(nil)

	// 2nd hash i.e. Hash of hash
	h.Reset()
	h.Write(sh1)
	sh2 := h.Sum(nil)

	return sh2[:addressChecksumLen]
}

// ValidatePublicKeyAddress validates the public key address returning true
// if it is valid
func ValidatePublicKeyAddress(address []byte, h hash.Hash) bool {

	pubKeyHash := base58.Decode(address)
	length := len(pubKeyHash) - addressChecksumLen
	actualChecksum := pubKeyHash[length:]

	pubKeyHash = pubKeyHash[:length]
	targetChecksum := checksum(pubKeyHash, h)

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
