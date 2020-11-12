package examples

import (
	"testing"

	"github.com/disintegration/imaging"
)

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
