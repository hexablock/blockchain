package hasher

import (
	"crypto/sha256"
	"hash"
)

// Hasher is a hash interface implemening a hash function
type Hasher interface {
	Clone() Hasher
	Name() string
	New() hash.Hash
	Size() int
}

// Default returns the default hasher
func Default() Hasher {
	return &SHA256Hasher{}
}

// New returns a new hash based on the given algorithm
func New(algo string) Hasher {
	switch algo {

	}
	return Default()
}

type SHA256Hasher struct{}

func (h *SHA256Hasher) Clone() Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) Name() string {
	return "sha256"
}
func (h *SHA256Hasher) New() hash.Hash {
	return sha256.New()
}
func (h *SHA256Hasher) Size() int {
	return sha256.Size
}
