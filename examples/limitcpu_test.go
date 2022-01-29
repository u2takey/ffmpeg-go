//+build linux

package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ComplexFilterExample(testInputFile, testOverlayFile, testOutputFile string) *ffmpeg.Stream {
	split := ffmpeg.Input(testInputFile).VFlip().Split()
	split0, split1 := split.Get("0"), split.Get("1")
	overlayFile := ffmpeg.Input(testOverlayFile).Crop(10, 10, 158, 112)
	return ffmpeg.Concat([]*ffmpeg.Stream{
		split0.Trim(ffmpeg.KwArgs{"start_frame": 10, "end_frame": 20}),
		split1.Trim(ffmpeg.KwArgs{"start_frame": 30, "end_frame": 40})}).
		Overlay(overlayFile.HFlip(), "").
		DrawBox(50, 50, 120, 120, "red", 5).
		Output(testOutputFile).
		OverWriteOutput()
}

// PID    USER       PR  NI    VIRT    RES    SHR S  %CPU   %MEM     TIME+ COMMAND
// 1386105 root      20   0 2114152 273780  31672 R  50.2   1.7      0:16.79 ffmpeg
func TestLimitCpu(t *testing.T) {
	e := ComplexFilterExample("./sample_data/in1.mp4", "./sample_data/overlay.png", "./sample_data/out2.mp4")
	err := e.RunWithResource(0.1, 0.5)
	if err != nil {
		assert.Nil(t, err)
	}
}
