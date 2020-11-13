# ffmpeg-go

ffmpeg-go is golang port of https://github.com/kkroening/ffmpeg-python

check ffmpeg_test.go for examples.


# example 

```go
split := Input(TestInputFile1).VFlip().Split()
	split0, split1 := split.Get("0"), split.Get("1")
	overlayFile := Input(TestOverlayFile).Crop(10, 10, 158, 112)
err := Concat([]*Stream{
    split0.Trim(KwArgs{"start_frame": 10, "end_frame": 20}),
    split1.Trim(KwArgs{"start_frame": 30, "end_frame": 40})}).
    Overlay(overlayFile.HFlip(), "").
    DrawBox(50, 50, 120, 120, "red", 5).
    Output(TestOutputFile1).
    OverWriteOutput().
    Run()
```

# view

function view generate [mermaid](https://mermaid-js.github.io/mermaid/#/) chart, which can be use in markdown or view [online](https://mermaid-js.github.io/mermaid-live-editor/)

```go
split := Input(TestInputFile1).VFlip().Split()
	split0, split1 := split.Get("0"), split.Get("1")
	overlayFile := Input(TestOverlayFile).Crop(10, 10, 158, 112)
b, err := Concat([]*Stream{
    split0.Trim(KwArgs{"start_frame": 10, "end_frame": 20}),
    split1.Trim(KwArgs{"start_frame": 30, "end_frame": 40})}).
    Overlay(overlayFile.HFlip(), "").
    DrawBox(50, 50, 120, 120, "red", 5).
    Output(TestOutputFile1).
    OverWriteOutput().View(ViewTypeFlowChart)
fmt.Println(b)
```
![image](./docs/flowchart2.png)