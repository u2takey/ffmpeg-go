package examples

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

//
// More simple examples please refer to ffmpeg_test.go
//

func TestExampleStream(t *testing.T) {
	ExampleStream("./sample_data/in1.mp4", "./sample_data/out1.mp4", false)
}

func TestExampleReadFrameAsJpeg(t *testing.T) {
	reader := ExampleReadFrameAsJpeg("./sample_data/in1.mp4", 5)
	img, err := imaging.Decode(reader)
	if err != nil {
		t.Fatal(err)
	}
	err = imaging.Save(img, "./sample_data/out1.jpeg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExampleReadTimePositionAsJpeg(t *testing.T) {
	reader := ExampleReadTimePositionAsJpeg("./sample_data/in1.mp4", 4)
	img, err := imaging.Decode(reader)
	if err != nil {
		t.Fatal(err)
	}
	err = imaging.Save(img, "./sample_data/out2.jpeg")
	if err != nil {
		t.Fatal(err)
	}
}

func TestExampleShowProgress(t *testing.T) {
	ExampleShowProgress("./sample_data/in1.mp4", "./sample_data/out2.mp4")
}

func TestSimpleS3StreamExample(t *testing.T) {
	err := ffmpeg.Input("./sample_data/in1.mp4", nil).
		Output("s3://data-1251825869/test_out.ts", ffmpeg.KwArgs{
			"aws_config": &aws.Config{
				Credentials: credentials.NewStaticCredentials("xx", "yyy", ""),
				//Endpoint:    aws.String("xx"),
				Region: aws.String("yyy"),
			},
			// outputS3 use stream output, so you can only use supported format
			// if you want mp4 format for example, you can output it to a file, and then call s3 sdk to do upload
			"format": "mpegts",
		}).
		Run()
	assert.Nil(t, err)
}

func TestExampleChangeCodec(t *testing.T) {
	err := ffmpeg.Input("./sample_data/in1.mp4").
		Output("./sample_data/out1.mp4", ffmpeg.KwArgs{"c:v": "libx265"}).
		OverWriteOutput().ErrorToStdOut().Run()
	assert.Nil(t, err)
}

func TestExampleCutVideo(t *testing.T) {
	err := ffmpeg.Input("./sample_data/in1.mp4", ffmpeg.KwArgs{"ss": 1}).
		Output("./sample_data/out1.mp4", ffmpeg.KwArgs{"t": 1}).OverWriteOutput().Run()
	assert.Nil(t, err)
}

func TestExampleScaleVideo(t *testing.T) {
	err := ffmpeg.Input("./sample_data/in1.mp4").
		Output("./sample_data/out1.mp4", ffmpeg.KwArgs{"vf": "scale=w=480:h=240"}).
		OverWriteOutput().ErrorToStdOut().Run()
	assert.Nil(t, err)
}

func TestExampleAddWatermark(t *testing.T) {
	// show watermark with size 64:-1 in the top left corner after seconds 1
	overlay := ffmpeg.Input("./sample_data/overlay.png").Filter("scale", ffmpeg.Args{"64:-1"})
	err := ffmpeg.Filter(
		[]*ffmpeg.Stream{
			ffmpeg.Input("./sample_data/in1.mp4"),
			overlay,
		}, "overlay", ffmpeg.Args{"10:10"}, ffmpeg.KwArgs{"enable": "gte(t,1)"}).
		Output("./sample_data/out1.mp4").OverWriteOutput().ErrorToStdOut().Run()
	assert.Nil(t, err)
}

func TestExampleCutVideoForGif(t *testing.T) {
	err := ffmpeg.Input("./sample_data/in1.mp4", ffmpeg.KwArgs{"ss": "1"}).
		Output("./sample_data/out1.gif", ffmpeg.KwArgs{"s": "320x240", "pix_fmt": "rgb24", "t": "3", "r": "3"}).
		OverWriteOutput().ErrorToStdOut().Run()
	assert.Nil(t, err)
}

func TestExampleMultipleOutput(t *testing.T) {
	input := ffmpeg.Input("./sample_data/in1.mp4").Split()
	out1 := input.Get("0").Filter("scale", ffmpeg.Args{"1920:-1"}).
		Output("./sample_data/1920.mp4", ffmpeg.KwArgs{"b:v": "5000k"})
	out2 := input.Get("1").Filter("scale", ffmpeg.Args{"1280:-1"}).
		Output("./sample_data/1280.mp4", ffmpeg.KwArgs{"b:v": "2800k"})

	err := ffmpeg.MergeOutputs(out1, out2).OverWriteOutput().ErrorToStdOut().Run()
	assert.Nil(t, err)
}
