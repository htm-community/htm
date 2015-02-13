package encoders

import (
	"github.com/stretchr/testify/assert"
	"github.com/zacg/htm/utils"
	"testing"
)

func TestEncoding(t *testing.T) {

	t.Log(h1)
	t.Log(h2)
	assert.True(t, h1 != h2)
	assert.True(t, h2 != h3)

}
