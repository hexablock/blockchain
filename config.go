package blockchain

import (
	"crypto/elliptic"

	"github.com/hexablock/hasher"
)

// Config holds the blockchain config
type Config struct {
	// Hash function to use
	Hasher hasher.Hasher

	// Elliptic curve for verification
	Curve elliptic.Curve

	// These need to be specified by the user and are required
	BlockStorage BlockStorage
	TxStorage    TxStorage
	DataKeyIndex DataKeyIndex
}

// DefaultConfig returns a config with the default hasher and elliptic curve
func DefaultConfig() *Config {
	return &Config{
		Hasher: hasher.Default(),
		Curve:  elliptic.P256(),
	}
}
