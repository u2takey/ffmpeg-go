package ffmpeg_go

import "time"

// for json spec

type GraphNode struct {
	Name          string   `json:"name"`
	InputStreams  []string `json:"input_streams"`
	OutputStreams []string `json:"output_streams"`
	Args          Args     `json:"args"`
	KwArgs        KwArgs   `json:"kw_args"`
}

type GraphOptions struct {
	Timeout         time.Duration
	OverWriteOutput bool
}

type Graph struct {
	OutputStream string       `json:"output_stream"`
	GraphOptions GraphOptions `json:"graph_options"`
	Nodes        []GraphNode  `json:"nodes"`
}
