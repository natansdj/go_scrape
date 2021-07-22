package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryEngine(t *testing.T) {
	var val int64

	memory := New()

	assert.Nil(t, memory.Init())

	memory.AddTotalCount(1)
	val = memory.GetTotalCount()
	assert.Equal(t, int64(1), val)

	memory.AddTotalCount(100)
	val = memory.GetTotalCount()
	assert.Equal(t, int64(101), val)

	// test reset db
	memory.Reset()
	val = memory.GetTotalCount()
	assert.Equal(t, int64(0), val)

	assert.NoError(t, memory.Close())
}
