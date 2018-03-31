package blockchain

import (
	"crypto/elliptic"

	"github.com/hexablock/blockchain/hasher"
)

// Config holds the blockchain config
type Config struct {
	Hasher         hasher.Hasher
	Curve          elliptic.Curve
	BlockValidator BlockValidator
	BlockStorage   BlockStorage
	TxStorage      TxStorage
	DataKeyIndex   DataKeyIndex
}

// DefaultConfig returns a config with the default hasher and elliptic curve
func DefaultConfig() *Config {
	return &Config{
		Hasher: hasher.Default(),
		Curve:  elliptic.P256(),
	}
}
