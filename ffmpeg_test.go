package ffmpeg_go

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/u2takey/go-utils/rand"
)

const (
	TestInputFile1  = "./examples/sample_data/in1.mp4"
	TestOutputFile1 = "./examples/sample_data/out1.mp4"
	TestOverlayFile = "./examples/sample_data/overlay.png"
)

func TestFluentEquality(t *testing.T) {
	base1 := Input("dummy1.mp4")
	base2 := Input("dummy1.mp4")
	base3 := Input("dummy2.mp4")
	t1 := base1.Trim(KwArgs{"start_frame": 10, "end_frame": 20})
	t2 := base1.Trim(KwArgs{"start_frame": 10, "end_frame": 20})
	t3 := base1.Trim(KwArgs{"start_frame": 10, "end_frame": 30})
	t4 := base2.Trim(KwArgs{"start_frame": 10, "end_frame": 20})
	t5 := base3.Trim(KwArgs{"start_frame": 10, "end_frame": 20})

	assert.Equal(t, t1.Hash(), t2.Hash())
	assert.Equal(t, t1.Hash(), t4.Hash())
	assert.NotEqual(t, t1.Hash(), t3.Hash())
	assert.NotEqual(t, t1.Hash(), t5.Hash())
}

func TestFluentConcat(t *testing.T) {
	base1 := Input("dummy1.mp4", nil)
	trim1 := base1.Trim(KwArgs{"start_frame": 10, "end_frame": 20})
	trim2 := base1.Trim(KwArgs{"start_frame": 30, "end_frame": 40})
	trim3 := base1.Trim(KwArgs{"start_frame": 50, "end_frame": 60})
	concat1 := Concat([]*Stream{trim1, trim2, trim3})
	concat2 := Concat([]*Stream{trim1, trim2, trim3})
	concat3 := Concat([]*Stream{trim1, trim3, trim2})
	assert.Equal(t, concat1.Hash(), concat2.Hash())
	assert.NotEqual(t, concat1.Hash(), concat3.Hash())
}

func TestRepeatArgs(t *testing.T) {
	o := Input("dummy.mp4", nil).Output("dummy2.mp4",
		KwArgs{"streamid": []string{"0:0x101", "1:0x102"}})
	assert.Equal(t, o.GetArgs(), []string{"-i", "dummy.mp4", "-streamid", "0:0x101", "-streamid", "1:0x102", "dummy2.mp4"})
}

func TestGlobalArgs(t *testing.T) {
	o := Input("dummy.mp4", nil).Output("dummy2.mp4", nil).GlobalArgs("-progress", "someurl")

	assert.Equal(t, o.GetArgs(), []string{
		"-i",
		"dummy.mp4",
		"dummy2.mp4",
		"-progress",
		"someurl",
	})
}

func TestSimpleExample(t *testing.T) {
	err := Input(TestInputFile1, nil).
		Output(TestOutputFile1, nil).
		OverWriteOutput().
		Run()
	assert.Nil(t, err)
}

func TestSimpleOverLayExample(t *testing.T) {
	err := Input(TestInputFile1, nil).
		Overlay(Input(TestOverlayFile), "").
		Output(TestOutputFile1).OverWriteOutput().
		Run()
	assert.Nil(t, err)
}

func TestSimpleOutputArgs(t *testing.T) {
	cmd := Input(TestInputFile1).Output("imageFromVideo_%d.jpg", KwArgs{"vf": "fps=3", "qscale:v": 2})
	assert.Equal(t, []string{
		"-i", "./examples/sample_data/in1.mp4", "-qscale:v",
		"2", "-vf", "fps=3", "imageFromVideo_%d.jpg"}, cmd.GetArgs())
}

func TestAutomaticStreamSelection(t *testing.T) {
	// example from http://ffmpeg.org/ffmpeg-all.html
	input := []*Stream{Input("A.avi"), Input("B.mp4")}
	out1 := Output(input, "out1.mkv")
	out2 := Output(input, "out2.wav")
	out3 := Output(input, "out3.mov", KwArgs{"map": "1:a", "c:a": "copy"})
	cmd := MergeOutputs(out1, out2, out3)
	printArgs(cmd.GetArgs())
	printGraph(cmd)
}

func TestLabeledFiltergraph(t *testing.T) {
	// example from http://ffmpeg.org/ffmpeg-all.html
	in1, in2, in3 := Input("A.avi"), Input("B.mp4"), Input("C.mkv")
	in2Split := in2.Get("v").Hue(KwArgs{"s": 0}).Split()
	overlay := Filter([]*Stream{in1, in2}, "overlay", nil)
	aresample := Filter([]*Stream{in1, in2, in3}, "aresample", nil)
	out1 := Output([]*Stream{in2Split.Get("outv1"), overlay, aresample}, "out1.mp4", KwArgs{"an": ""})
	out2 := Output([]*Stream{in1, in2, in3}, "out2.mkv")
	out3 := in2Split.Get("outv2").Output("out3.mkv", KwArgs{"map": "1:a:0"})
	cmd := MergeOutputs(out1, out2, out3)
	printArgs(cmd.GetArgs())
	printGraph(cmd)
}

func ComplexFilterExample() *Stream {
	split := Input(TestInputFile1).VFlip().Split()
	split0, split1 := split.Get("0"), split.Get("1")
	overlayFile := Input(TestOverlayFile).Crop(10, 10, 158, 112)
	return Concat([]*Stream{
		split0.Trim(KwArgs{"start_frame": 10, "end_frame": 20}),
		split1.Trim(KwArgs{"start_frame": 30, "end_frame": 40})}).
		Overlay(overlayFile.HFlip(), "").
		DrawBox(50, 50, 120, 120, "red", 5).
		Output(TestOutputFile1).
		OverWriteOutput()
}

func TestComplexFilterExample(t *testing.T) {
	assert.Equal(t, []string{
		"-i",
		TestInputFile1,
		"-i",
		TestOverlayFile,
		"-filter_complex",
		"[0]vflip[s0];" +
			"[s0]split=2[s1][s2];" +
			"[s1]trim=end_frame=20:start_frame=10[s3];" +
			"[s2]trim=end_frame=40:start_frame=30[s4];" +
			"[s3][s4]concat=n=2[s5];" +
			"[1]crop=158:112:10:10[s6];" +
			"[s6]hflip[s7];" +
			"[s5][s7]overlay=eof_action=repeat[s8];" +
			"[s8]drawbox=50:50:120:120:red:t=5[s9]",
		"-map",
		"[s9]",
		TestOutputFile1,
		"-y",
	}, ComplexFilterExample().GetArgs())
}

func TestCombinedOutput(t *testing.T) {
	i1 := Input(TestInputFile1)
	i2 := Input(TestOverlayFile)
	out := Output([]*Stream{i1, i2}, TestOutputFile1)
	assert.Equal(t, []string{
		"-i",
		TestInputFile1,
		"-i",
		TestOverlayFile,
		"-map",
		"0",
		"-map",
		"1",
		TestOutputFile1,
	}, out.GetArgs())
}

func TestFilterWithSelector(t *testing.T) {
	i := Input(TestInputFile1)

	v1 := i.Video().HFlip()
	a1 := i.Audio().Filter("aecho", Args{"0.8", "0.9", "1000", "0.3"})

	out := Output([]*Stream{a1, v1}, TestOutputFile1)
	assert.Equal(t, []string{
		"-i",
		TestInputFile1,
		"-filter_complex",
		"[0:a]aecho=0.8:0.9:1000:0.3[s0];[0:v]hflip[s1]",
		"-map",
		"[s0]",
		"-map",
		"[s1]",
		TestOutputFile1}, out.GetArgs())

}

func ComplexFilterAsplitExample() *Stream {
	split := Input(TestInputFile1).VFlip().ASplit()
	split0 := split.Get("0")
	split1 := split.Get("1")

	return Concat([]*Stream{
		split0.Filter("atrim", nil, KwArgs{"start": 10, "end": 20}),
		split1.Filter("atrim", nil, KwArgs{"start": 30, "end": 40}),
	}).Output(TestOutputFile1).OverWriteOutput()
}

func TestFilterConcatVideoOnly(t *testing.T) {
	in1 := Input("in1.mp4")
	in2 := Input("in2.mp4")
	args := Concat([]*Stream{in1, in2}).Output("out.mp4").GetArgs()
	assert.Equal(t, []string{
		"-i",
		"in1.mp4",
		"-i",
		"in2.mp4",
		"-filter_complex",
		"[0][1]concat=n=2[s0]",
		"-map",
		"[s0]",
		"out.mp4",
	}, args)
}

func TestFilterConcatAudioOnly(t *testing.T) {
	in1 := Input("in1.mp4")
	in2 := Input("in2.mp4")
	args := Concat([]*Stream{in1, in2}, KwArgs{"v": 0, "a": 1}).Output("out.mp4").GetArgs()
	assert.Equal(t, []string{
		"-i",
		"in1.mp4",
		"-i",
		"in2.mp4",
		"-filter_complex",
		"[0][1]concat=a=1:n=2:v=0[s0]",
		"-map",
		"[s0]",
		"out.mp4",
	}, args)
}

func TestFilterConcatAudioVideo(t *testing.T) {
	in1 := Input("in1.mp4")
	in2 := Input("in2.mp4")
	joined := Concat([]*Stream{in1.Video(), in1.Audio(), in2.HFlip(), in2.Get("a")}, KwArgs{"v": 1, "a": 1}).Node
	args := Output([]*Stream{joined.Get("0"), joined.Get("1")}, "out.mp4").GetArgs()
	assert.Equal(t, []string{
		"-i",
		"in1.mp4",
		"-i",
		"in2.mp4",
		"-filter_complex",
		"[1]hflip[s0];[0:v][0:a][s0][1:a]concat=a=1:n=2:v=1[s1][s2]",
		"-map",
		"[s1]",
		"-map",
		"[s2]",
		"out.mp4",
	}, args)
}

func TestFilterASplit(t *testing.T) {
	out := ComplexFilterAsplitExample()
	args := out.GetArgs()
	assert.Equal(t, []string{
		"-i",
		TestInputFile1,
		"-filter_complex",
		"[0]vflip[s0];[s0]asplit=2[s1][s2];[s1]atrim=end=20:start=10[s3];[s2]atrim=end=40:start=30[s4];[s3][s4]concat=n=2[s5]",
		"-map",
		"[s5]",
		TestOutputFile1,
		"-y",
	}, args)
}

func TestOutputBitrate(t *testing.T) {
	args := Input("in").Output("out", KwArgs{"video_bitrate": 1000, "audio_bitrate": 200}).GetArgs()
	assert.Equal(t, []string{"-i", "in", "-b:v", "1000", "-b:a", "200", "out"}, args)
}

func TestOutputVideoSize(t *testing.T) {
	args := Input("in").Output("out", KwArgs{"video_size": "320x240"}).GetArgs()
	assert.Equal(t, []string{"-i", "in", "-video_size", "320x240", "out"}, args)
}

func TestCompile(t *testing.T) {
	out := Input("dummy.mp4").Output("dummy2.mp4")
	assert.Equal(t, out.Compile().Args, []string{"ffmpeg", "-i", "dummy.mp4", "dummy2.mp4"})
}

func TestPipe(t *testing.T) {

	width, height := 32, 32
	frameSize := width * height * 3
	frameCount, startFrame := 10, 2
	_, _ = frameCount, frameSize

	out := Input(
		"pipe:0",
		KwArgs{
			"format":       "rawvideo",
			"pixel_format": "rgb24",
			"video_size":   fmt.Sprintf("%dx%d", width, height),
			"framerate":    10}).
		Trim(KwArgs{"start_frame": startFrame}).
		Output("pipe:1", KwArgs{"format": "rawvideo"})

	args := out.GetArgs()
	assert.Equal(t, args, []string{
		"-f",
		"rawvideo",
		"-video_size",
		fmt.Sprintf("%dx%d", width, height),
		"-framerate",
		"10",
		"-pixel_format",
		"rgb24",
		"-i",
		"pipe:0",
		"-filter_complex",
		"[0]trim=start_frame=2[s0]",
		"-map",
		"[s0]",
		"-f",
		"rawvideo",
		"pipe:1",
	})

	inBuf := bytes.NewBuffer(nil)
	for i := 0; i < frameSize*frameCount; i++ {
		inBuf.WriteByte(byte(rand.IntnRange(0, 255)))
	}
	outBuf := bytes.NewBuffer(nil)
	err := out.WithInput(inBuf).WithOutput(outBuf).Run()
	assert.Nil(t, err)
	assert.Equal(t, outBuf.Len(), frameSize*(frameCount-startFrame))
}

func TestView(t *testing.T) {
	a, err := ComplexFilterExample().View(ViewTypeFlowChart)
	assert.Nil(t, err)

	b, err := ComplexFilterAsplitExample().View(ViewTypeStateDiagram)
	assert.Nil(t, err)

	t.Log(a)
	t.Log(b)
}

func printArgs(args []string) {
	for _, a := range args {
		fmt.Printf("%s ", a)
	}
	fmt.Println()
}

func printGraph(s *Stream) {
	fmt.Println()
	v, _ := s.View(ViewTypeFlowChart)
	fmt.Println(v)
}

//func TestAvFoundation(t *testing.T) {
//	out := Input("default:none", KwArgs{"f": "avfoundation", "framerate": "30"}).
//		Output("output.mp4", KwArgs{"format": "mp4"}).
//		OverWriteOutput()
//	assert.Equal(t, []string{"-f", "avfoundation", "-framerate",
//		"30", "-i", "default:none", "-f", "mp4", "output.mp4", "-y"}, out.GetArgs())
//	err := out.Run()
//	assert.Nil(t, err)
//}
