package bcpb

// HasPublicKey return true if the public key is one of the public keys
// listed in the output
func (txo *TxOutput) HasPublicKey(pk PublicKey) bool {
	if len(txo.PubKeys) == 0 {
		return true
	}

	for _, p := range txo.PubKeys {
		if p.Equal(p) {
			return true
		}
	}

	return false
}

// RemovePublicKey removes the public key returning true if it was removed
func (txo *TxOutput) RemovePublicKey(pk PublicKey) bool {
	for i := range txo.PubKeys {
		if txo.PubKeys[i].Equal(pk) {
			txo.PubKeys = append(txo.PubKeys[0:i], txo.PubKeys[i+1:]...)
			return true
		}
	}
	return false
}

// SetRequiredSignatures sets the required signatures to mutate the output
func (txo *TxOutput) SetRequiredSignatures(c uint8) {
	if len(txo.Logic) > 0 {
		txo.Logic[0] = c
	} else {
		txo.Logic = []byte{c}
	}
}

// Copy returns a copy of the transaction output
func (txo *TxOutput) Copy() *TxOutput {
	o := &TxOutput{
		DataKey: txo.DataKey,
		Metrics: make(map[string]float64, len(txo.Metrics)),
		Tags:    make(map[string]string, len(txo.Tags)),
		Labels:  make([]string, len(txo.Labels)),
		PubKeys: make([]PublicKey, len(txo.PubKeys)),
		Logic:   make([]byte, len(txo.Logic)),
	}

	for k, v := range txo.Metrics {
		o.Metrics[k] = v
	}

	for k, v := range txo.Tags {
		o.Tags[k] = v
	}

	copy(o.Labels, txo.Labels)

	for i := range txo.PubKeys {
		o.PubKeys[i] = txo.PubKeys[i]
	}

	copy(o.Logic, txo.Logic)

	return o
}
