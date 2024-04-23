package examples

import (
	"bytes"
	"io"
	"os"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ExampleReadTimePositionAsJpeg(inFileName string, seconds int) io.Reader {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(inFileName, ffmpeg.KwArgs{"ss": seconds}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}
