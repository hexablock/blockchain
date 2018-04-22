package bcpb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DataKey(t *testing.T) {
	dk1 := ParseDataKey("foo:bar")
	assert.Equal(t, "foo", string(dk1.Type()))

	dk2 := ParseDataKey("foo")
	assert.Equal(t, "foo", string(dk2.Type()))
	assert.Equal(t, "", string(dk2.ID()))

	dk3 := ParseDataKey("foo:")
	assert.Equal(t, "", string(dk3.ID()))

	dk4 := ParseDataKey("foo:bar:baz")
	assert.Equal(t, "bar:baz", string(dk4.ID()))

	dk5 := ParseDataKey("")
	assert.Equal(t, "", string(dk5.Type()))
	assert.Equal(t, "", string(dk5.ID()))
}
