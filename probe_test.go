package ffmpeg_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

const (
	TestInputFile1  = "./sample_data/in1.mp4"
	TestOutputFile1 = "./sample_data/out1.mp4"
	TestOverlayFile = "./sample_data/overlay.png"
)

func TestProbe(t *testing.T) {
	data, err := Probe(TestInputFile1, nil)
	assert.Nil(t, err)
	assert.Equal(t, gjson.Get(data, "format.duration").String(), "7.036000")
}
