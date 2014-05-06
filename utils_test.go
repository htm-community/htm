package htm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCartProductInt(t *testing.T) {
	vals := [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{10, 11, 12, 13},
	}

	result := CartProductInt(vals)

}
