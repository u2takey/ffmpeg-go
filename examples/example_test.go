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
