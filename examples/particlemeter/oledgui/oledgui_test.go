package oledgui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAxis(t *testing.T) {
	//step := 0.09

	//s0 := fmt.Sprintf(decimalFormatString(step), step)
	//s1 := fmt.Sprintf(decimalFormatString(step), step*2)

	assert.Equal(t, "%.1f", decimalFormatString(0.1))
	assert.Equal(t, "%.2f", decimalFormatString(0.01))
	assert.Equal(t, "%.3f", decimalFormatString(0.001))
	assert.Equal(t, "%.4f", decimalFormatString(0.0001))

	assert.Equal(t, "%.1f", decimalFormatString(1.1))
	assert.Equal(t, "%.1f", decimalFormatString(11.1))
	assert.Equal(t, "%.1f", decimalFormatString(111.1))
	assert.Equal(t, "%.1f", decimalFormatString(1111.1))
	assert.Equal(t, "%.1f", decimalFormatString(11111.1))
	assert.Equal(t, "%.3f", decimalFormatString(11111.001))
	assert.Equal(t, "%.0f", decimalFormatString(5))
	//t.Errorf("s0=%s s1=%s\n", s0, s1)

}
