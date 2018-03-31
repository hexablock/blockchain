package bcpb

import "bytes"

// DataKey represents a data key. This is a token key as opposed to a crypto key
// It is in the format <type>:<some uinque key in type>, where type is the
// datatype with the remainder to be used as an identifier
type DataKey []byte

// NewDataKey returns a new data key with the given type and id delimited by ':'
func NewDataKey(typ, key []byte) DataKey {
	return DataKey(append(append(typ, ':'), key...))
}

func (k DataKey) String() string {
	return string(k)
}

// Equal trues true if both keys are equal
func (k DataKey) Equal(b []byte) bool {
	return bytes.Compare(k, b) == 0
}

// Type returns the key type par of the key. This is all bytes before the first '/'
func (k DataKey) Type() []byte {
	i := bytes.IndexByte(k, ':')
	if i < 0 {
		return k
	}
	return k[:i]
}

// ID returns the id part of the key and nil if the key does not contain an id
func (k DataKey) ID() []byte {
	i := bytes.IndexByte(k, ':')
	if i < 0 {
		return nil
	}
	return k[i+1:]
}
