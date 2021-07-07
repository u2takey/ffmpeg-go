package ffmpeg_go

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func getInputArgs(node *Node) []string {
	var args []string
	if node.name == "input" {
		kwargs := node.kwargs.Copy()
		filename := kwargs.PopString("filename")
		format := kwargs.PopString("format")
		videoSize := kwargs.PopString("video_size")
		if format != "" {
			args = append(args, "-f", format)
		}
		if videoSize != "" {
			args = append(args, "-video_size", videoSize)
		}
		args = append(args, ConvertKwargsToCmdLineArgs(kwargs)...)
		args = append(args, "-i", filename)
	} else {
		panic("unsupported node input name")
	}
	return args
}

func formatInputStreamName(streamNameMap map[string]string, edge DagEdge, finalArg bool) string {
	prefix := streamNameMap[fmt.Sprintf("%d%s", edge.UpStreamNode.Hash(), edge.UpStreamLabel)]
	suffix := ""
	format := "[%s%s]"
	if edge.UpStreamSelector != "" {
		suffix = fmt.Sprintf(":%s", edge.UpStreamSelector)
	}
	if finalArg && edge.UpStreamNode.(*Node).nodeType == "InputNode" {
		format = "%s%s"
	}
	return fmt.Sprintf(format, prefix, suffix)
}

func formatOutStreamName(streamNameMap map[string]string, edge DagEdge) string {
	return fmt.Sprintf("[%s]", streamNameMap[fmt.Sprintf("%d%s", edge.UpStreamNode.Hash(), edge.UpStreamLabel)])
}

func _getFilterSpec(node *Node, outOutingEdgeMap map[Label][]NodeInfo, streamNameMap map[string]string) string {
	var input, output []string
	for _, e := range node.GetInComingEdges() {
		input = append(input, formatInputStreamName(streamNameMap, e, false))
	}
	outEdges := GetOutGoingEdges(node, outOutingEdgeMap)
	for _, e := range outEdges {
		output = append(output, formatOutStreamName(streamNameMap, e))
	}
	return fmt.Sprintf("%s%s%s", strings.Join(input, ""), node.GetFilter(outEdges), strings.Join(output, ""))
}

func _getAllLabelsInSorted(m map[Label]NodeInfo) []Label {
	var r []Label
	for a := range m {
		r = append(r, a)
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})
	return r
}

func _getAllLabelsSorted(m map[Label][]NodeInfo) []Label {
	var r []Label
	for a := range m {
		r = append(r, a)
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i] < r[j]
	})
	return r
}

func _allocateFilterStreamNames(nodes []*Node, outOutingEdgeMaps map[int]map[Label][]NodeInfo, streamNameMap map[string]string) {
	sc := 0
	for _, n := range nodes {
		om := outOutingEdgeMaps[n.Hash()]
		// todo sort
		for _, l := range _getAllLabelsSorted(om) {
			if len(om[l]) > 1 {
				panic(fmt.Sprintf(`encountered %s with multiple outgoing edges 
with same upstream label %s; a 'split'' filter is probably required`, n.name, l))
			}
			streamNameMap[fmt.Sprintf("%d%s", n.Hash(), l)] = fmt.Sprintf("s%d", sc)
			sc += 1
		}
	}
}

func _getFilterArg(nodes []*Node, outOutingEdgeMaps map[int]map[Label][]NodeInfo, streamNameMap map[string]string) string {
	_allocateFilterStreamNames(nodes, outOutingEdgeMaps, streamNameMap)
	var filterSpec []string
	for _, n := range nodes {
		filterSpec = append(filterSpec, _getFilterSpec(n, outOutingEdgeMaps[n.Hash()], streamNameMap))
	}
	return strings.Join(filterSpec, ";")
}

func _getGlobalArgs(node *Node) []string {
	return node.args
}

func _getOutputArgs(node *Node, streamNameMap map[string]string) []string {
	if node.name != "output" {
		panic("Unsupported output node")
	}
	var args []string
	if len(node.GetInComingEdges()) == 0 {
		panic("Output node has no mapped streams")
	}
	for _, e := range node.GetInComingEdges() {
		streamName := formatInputStreamName(streamNameMap, e, true)
		if streamName != "0" || len(node.GetInComingEdges()) > 1 {
			args = append(args, "-map", streamName)
		}
	}
	kwargs := node.kwargs.Copy()

	filename := kwargs.PopString("filename")
	if kwargs.HasKey("format") {
		args = append(args, "-f", kwargs.PopString("format"))
	}
	if kwargs.HasKey("video_bitrate") {
		args = append(args, "-b:v", kwargs.PopString("video_bitrate"))
	}
	if kwargs.HasKey("audio_bitrate") {
		args = append(args, "-b:a", kwargs.PopString("audio_bitrate"))
	}
	if kwargs.HasKey("video_size") {
		args = append(args, "-video_size", kwargs.PopString("video_size"))
	}

	args = append(args, ConvertKwargsToCmdLineArgs(kwargs)...)
	args = append(args, filename)
	return args
}

func (s *Stream) GetArgs() []string {
	var args []string
	nodes := getStreamSpecNodes([]*Stream{s})
	var dagNodes []DagNode
	streamNameMap := map[string]string{}
	for i := range nodes {
		dagNodes = append(dagNodes, nodes[i])
	}
	sorted, outGoingMap, err := TopSort(dagNodes)
	if err != nil {
		panic(err)
	}
	DebugNodes(sorted)
	DebugOutGoingMap(sorted, outGoingMap)
	var inputNodes, outputNodes, globalNodes, filterNodes []*Node
	for i := range sorted {
		n := sorted[i].(*Node)
		switch n.nodeType {
		case "InputNode":
			streamNameMap[fmt.Sprintf("%d", n.Hash())] = fmt.Sprintf("%d", len(inputNodes))
			inputNodes = append(inputNodes, n)
		case "OutputNode":
			outputNodes = append(outputNodes, n)
		case "GlobalNode":
			globalNodes = append(globalNodes, n)
		case "FilterNode":
			filterNodes = append(filterNodes, n)
		}
	}
	// input args from inputNodes
	for _, n := range inputNodes {
		args = append(args, getInputArgs(n)...)
	}
	// filter args from filterNodes
	filterArgs := _getFilterArg(filterNodes, outGoingMap, streamNameMap)
	if filterArgs != "" {
		args = append(args, "-filter_complex", filterArgs)
	}
	// output args from outputNodes
	for _, n := range outputNodes {
		args = append(args, _getOutputArgs(n, streamNameMap)...)
	}
	// global args with outputNodes
	for _, n := range globalNodes {
		args = append(args, _getGlobalArgs(n)...)
	}
	if s.Context.Value("OverWriteOutput") != nil {
		args = append(args, "-y")
	}
	return args
}

func (s *Stream) WithTimeout(timeOut time.Duration) *Stream {
	if timeOut > 0 {
		s.Context, _ = context.WithTimeout(s.Context, timeOut)
	}
	return s
}

func (s *Stream) OverWriteOutput() *Stream {
	s.Context = context.WithValue(s.Context, "OverWriteOutput", struct{}{})
	return s
}

func (s *Stream) WithInput(reader io.Reader) *Stream {
	s.Context = context.WithValue(s.Context, "Stdin", reader)
	return s
}

func (s *Stream) WithOutput(out ...io.Writer) *Stream {
	if len(out) > 0 {
		s.Context = context.WithValue(s.Context, "Stdout", out[0])
	}
	if len(out) > 1 {
		s.Context = context.WithValue(s.Context, "Stderr", out[1])
	}
	return s
}

func (s *Stream) WithErrorOutput(out io.Writer) *Stream {
	s.Context = context.WithValue(s.Context, "Stderr", out)
	return s
}

func (s *Stream) ErrorToStdOut() *Stream {
	return s.WithErrorOutput(os.Stdout)
}

// for test
func (s *Stream) Compile() *exec.Cmd {
	args := s.GetArgs()
	cmd := exec.CommandContext(s.Context, "ffmpeg", args...)
	if a, ok := s.Context.Value("Stdin").(io.Reader); ok {
		cmd.Stdin = a
	}
	if a, ok := s.Context.Value("Stdout").(io.Writer); ok {
		cmd.Stdout = a
	}
	if a, ok := s.Context.Value("Stderr").(io.Writer); ok {
		cmd.Stderr = a
	}
	log.Printf("compiled command: ffmpeg %s\n", strings.Join(args, " "))
	return cmd
}

func (s *Stream) Run() error {
	if s.Context.Value("run_hook") != nil {
		hook := s.Context.Value("run_hook").(*RunHook)
		go hook.f()
		defer func() {
			if hook.closer != nil {
				_ = hook.closer.Close()
			}
			<-hook.done
		}()
	}
	return s.Compile().Run()
}
