package htm

import (
	"github.com/zacg/testify/assert"
	"testing"
)

func TestNumColumns(t *testing.T) {
	c := NewTemporalMemoryConnections(0, 32, []int{64, 64})
	assert.Equal(t, 64*64, c.NumberOfColumns())
}
