package ffmpeg_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestProbe(t *testing.T) {
	data, err := Probe(TestInputFile1, nil)
	assert.Nil(t, err)
	assert.Equal(t, gjson.Get(data, "format.duration").String(), "7.036000")
}
