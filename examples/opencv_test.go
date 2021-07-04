// +build gocv

// uncomment line above for gocv examples

package examples

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"testing"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gocv.io/x/gocv"
)

// TestExampleOpenCvFaceDetect will: take a video as input => use opencv for face detection => draw box and show a window
// This example depends on gocv and opencv, please refer: https://pkg.go.dev/gocv.io/x/gocv for installation.
func TestExampleOpenCvFaceDetectWithVideo(t *testing.T) {
	inputFile := "./sample_data/head-pose-face-detection-male-short.mp4"
	xmlFile := "./sample_data/haarcascade_frontalface_default.xml"

	w, h := getVideoSize(inputFile)
	log.Println(w, h)

	pr1, pw1 := io.Pipe()
	readProcess(inputFile, pw1)
	openCvProcess(xmlFile, pr1, w, h)
	log.Println("Done")
}

func readProcess(infileName string, writer io.WriteCloser) {
	log.Println("Starting ffmpeg process1")
	go func() {
		err := ffmpeg.Input(infileName).
			Output("pipe:",
				ffmpeg.KwArgs{
					"format": "rawvideo", "pix_fmt": "rgb24",
				}).
			WithOutput(writer).
			ErrorToStdOut().
			Run()
		log.Println("ffmpeg process1 done")
		_ = writer.Close()
		if err != nil {
			panic(err)
		}
	}()
	return
}

func openCvProcess(xmlFile string, reader io.ReadCloser, w, h int) {
	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()

	// color for the rect when faces detected
	blue := color.RGBA{B: 255}

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
		img2 := gocv.NewMat()
		gocv.CvtColor(img, &img2, gocv.ColorBGRToRGB)

		// detect faces
		rects := classifier.DetectMultiScale(img2)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image, along with text identifing as "Human"
		for _, r := range rects {
			gocv.Rectangle(&img2, r, blue, 3)

			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img2, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img2)
		img.Close()
		img2.Close()
		if window.WaitKey(10) >= 0 {
			break
		}
	}
	return
}

// TestExampleOpenCvFaceDetectWithCamera will: task stream from webcam => use opencv for face detection => output with ffmpeg
// This example depends on gocv and opencv, please refer: https://pkg.go.dev/gocv.io/x/gocv for installation.
func TestExampleOpenCvFaceDetectWithCamera(t *testing.T) {
	deviceID := "0" // camera device id
	xmlFile := "./sample_data/haarcascade_frontalface_default.xml"

	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); !ok {
		panic(fmt.Sprintf("Cannot read device %v", deviceID))
	}
	fmt.Printf("img: %vX%v\n", img.Cols(), img.Rows())

	pr1, pw1 := io.Pipe()
	writeProcess("./sample_data/face_detect.mp4", pr1, img.Cols(), img.Rows())

	// color for the rect when faces detected
	blue := color.RGBA{B: 255}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	fmt.Printf("Start reading device: %v\n", deviceID)
	for i := 0; i < 200; i++ {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image, along with text identifing as "Human"
		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)

			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}
		pw1.Write(img.ToBytes())
	}
	pw1.Close()
	log.Println("Done")
}

func writeProcess(outputFile string, reader io.ReadCloser, w, h int) {
	log.Println("Starting ffmpeg process1")
	go func() {
		err := ffmpeg.Input("pipe:",
			ffmpeg.KwArgs{"format": "rawvideo",
				"pix_fmt": "bgr24", "s": fmt.Sprintf("%dx%d", w, h),
			}).
			Overlay(ffmpeg.Input("./sample_data/overlay.png"), "").
			Output(outputFile).
			WithInput(reader).
			ErrorToStdOut().
			OverWriteOutput().
			Run()
		log.Println("ffmpeg process1 done")
		if err != nil {
			panic(err)
		}
		_ = reader.Close()
	}()
}
