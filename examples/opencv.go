package examples

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gocv.io/x/gocv"
)

func readProcess(infileName string, writer io.WriteCloser) <-chan error {
	log.Println("Starting ffmpeg process1")
	done := make(chan error)
	go func() {
		err := ffmpeg.Input(infileName).
			Output("pipe:",
				ffmpeg.KwArgs{
					"format": "rawvideo", "pix_fmt": "rgb24",
				}).
			WithOutput(writer).
			Run()
		log.Println("ffmpeg process1 done")
		_ = writer.Close()
		done <- err
		close(done)
	}()
	return done
}

func openCvProcess(xmlFile string, reader io.ReadCloser, w, h int) {
	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	frameSize := w * h * 3
	buf := make([]byte, frameSize, frameSize)
	for {
		n, err := io.ReadFull(reader, buf)
		if n == 0 || err == io.EOF {
			return
		} else if n != frameSize || err != nil {
			panic(fmt.Sprintf("read error: %d, %s", n, err))
		}
		img, err := gocv.NewMatFromBytes(h, w, gocv.MatTypeCV8UC3, buf)
		if err != nil {
			fmt.Println("decode fail", err)
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image,
		// along with text identifing as "Human"
		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)

			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(10) >= 0 {
			break
		}
	}
	return
}

func ExampleFaceDetection(inputFile, xmlFile string) {
	w, h := getVideoSize(inputFile)
	log.Println(w, h)

	pr1, pw1 := io.Pipe()
	done1 := readProcess(inputFile, pw1)
	openCvProcess(xmlFile, pr1, w, h)
	err := <-done1
	if err != nil {
		panic(err)
	}
	log.Println("Done")
}
