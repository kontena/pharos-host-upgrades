package proc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadStat(t *testing.T) {
	stat, err := ReadStat()

	assert.NoErrorf(t, err, "ReadStat")
	assert.NotZerof(t, stat.BootTime.Unix(), "Stat.BootTime")
}
