package examples

import (
	"encoding/json"
	"fmt"
	"io"
	"log"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// ExampleStream
// inFileName: input filename
// outFileName: output filename
// dream: Use DeepDream frame processing (requires tensorflow)
func ExampleStream(inFileName, outFileName string, dream bool) {
	if inFileName == "" {
		inFileName = "./in1.mp4"
	}
	if outFileName == "" {
		outFileName = "./out.mp4"
	}
	if dream {
		panic("Use DeepDream With Tensorflow haven't been implemented")
	}

	runExampleStream(inFileName, outFileName)
}

func getVideoSize(fileName string) (int, int) {
	log.Println("Getting video size for", fileName)
	data, err := ffmpeg.Probe(fileName)
	if err != nil {
		panic(err)
	}
	log.Println("got video info", data)
	type VideoInfo struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int
			Height    int
		} `json:"streams"`
	}
	vInfo := &VideoInfo{}
	err = json.Unmarshal([]byte(data), vInfo)
	if err != nil {
		panic(err)
	}
	for _, s := range vInfo.Streams {
		if s.CodecType == "video" {
			return s.Width, s.Height
		}
	}
	return 0, 0
}

func startFFmpegProcess1(infileName string, writer io.WriteCloser) <-chan error {
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

func startFFmpegProcess2(outfileName string, buf io.Reader, width, height int) <-chan error {
	log.Println("Starting ffmpeg process2")
	done := make(chan error)
	go func() {
		err := ffmpeg.Input("pipe:",
			ffmpeg.KwArgs{"format": "rawvideo",
				"pix_fmt": "rgb24", "s": fmt.Sprintf("%dx%d", width, height),
			}).
			Output(outfileName, ffmpeg.KwArgs{"pix_fmt": "yuv420p"}).
			OverWriteOutput().
			WithInput(buf).
			Run()
		log.Println("ffmpeg process2 done")
		done <- err
		close(done)
	}()
	return done
}

func process(reader io.ReadCloser, writer io.WriteCloser, w, h int) {
	go func() {
		frameSize := w * h * 3
		buf := make([]byte, frameSize, frameSize)
		for {
			n, err := io.ReadFull(reader, buf)
			if n == 0 || err == io.EOF {
				_ = writer.Close()
				return
			} else if n != frameSize || err != nil {
				panic(fmt.Sprintf("read error: %d, %s", n, err))
			}
			for i := range buf {
				buf[i] = buf[i] / 3
			}
			n, err = writer.Write(buf)
			if n != frameSize || err != nil {
				panic(fmt.Sprintf("write error: %d, %s", n, err))
			}
		}
	}()
	return
}

func runExampleStream(inFile, outFile string) {
	w, h := getVideoSize(inFile)
	log.Println(w, h)

	pr1, pw1 := io.Pipe()
	pr2, pw2 := io.Pipe()
	done1 := startFFmpegProcess1(inFile, pw1)
	process(pr1, pw2, w, h)
	done2 := startFFmpegProcess2(outFile, pr2, w, h)
	err := <-done1
	if err != nil {
		panic(err)
	}
	err = <-done2
	if err != nil {
		panic(err)
	}
	log.Println("Done")
}
