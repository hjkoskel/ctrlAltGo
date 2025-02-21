package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKernelModuleState(t *testing.T) {
	a, err := ParseKernelModuleState("kfifo_buf 12288 0 - Live 0xffffffe126ab2000")
	assert.Equal(t, nil, err)
	assert.Equal(t, a, KernelModuleState{
		Name:       "kfifo_buf",
		Size:       12288,
		Instances:  0,
		UsedBy:     []string{},
		State:      LIVE,
		Address:    0xffffffe126ab2000,
		Annotation: "",
	})

}
