package filters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEqualFilter(t *testing.T) {
	// Test Equal
	filter1 := Equal(5)
	assert.True(t, filter1.Check(5), "Equal filter should return true if value equals")
	assert.False(t, filter1.Check(4), "Equal filter should return false if value does not equal")

	// Test NotEqual
	filter2 := NotEqual(5.1)
	assert.False(t, filter2.Check(5.1), "NotEqual filter should return false if value equals")
	assert.True(t, filter2.Check(4.1), "NotEqual filter should return true if value does not equal")
}

func TestInFilter(t *testing.T) {
	// Test In
	filter1 := In([]int{1, 2, 3})
	assert.True(t, filter1.Check(1), "In filter should return true if value is in the set")
	assert.False(t, filter1.Check(4), "In filter should return false if value is not in the set")

	// Test NotIn
	filter2 := NotIn([]int{1, 2, 3})
	assert.False(t, filter2.Check(1), "NotIn filter should return false if value is in the set")
	assert.True(t, filter2.Check(4), "NotIn filter should return true if value is not in the set")
}
