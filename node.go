package ffmpeg_go

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/u2takey/go-utils/sets"
)

type Stream struct {
	Node     *Node
	Label    Label
	Selector Selector
	Type     string
	Context  context.Context
}

type RunHook struct {
	f      func()
	done   <-chan struct{}
	closer interface {
		Close() error
	}
}

func NewStream(node *Node, streamType string, label Label, selector Selector) *Stream {
	return &Stream{
		Node:     node,
		Label:    label,
		Selector: selector,
		Type:     streamType,
		Context:  context.Background(),
	}
}

func (s *Stream) Hash() int {
	return s.Node.Hash() + getHash(s.Label)
}

func (s *Stream) Equal(other Stream) bool {
	return s.Hash() == other.Hash()
}

func (s *Stream) String() string {
	return fmt.Sprintf("node: %s, label: %s, selector: %s", s.Node.String(), s.Label, s.Selector)
}

func (s *Stream) Get(index string) *Stream {
	if s.Selector != "" {
		panic(errors.New("stream already has a selector"))
	}
	return s.Node.Stream(s.Label, Selector(index))
}

func (s *Stream) Audio() *Stream {
	return s.Get("a")
}

func (s *Stream) Video() *Stream {
	return s.Get("v")
}

func getStreamMap(streamSpec []*Stream) map[int]*Stream {
	m := map[int]*Stream{}
	for i := range streamSpec {
		m[i] = streamSpec[i]
	}
	return m
}

func getStreamMapNodes(streamMap map[int]*Stream) (ret []*Node) {
	for k := range streamMap {
		ret = append(ret, streamMap[k].Node)
	}
	return ret
}

func getStreamSpecNodes(streamSpec []*Stream) []*Node {
	return getStreamMapNodes(getStreamMap(streamSpec))
}

type Node struct {
	streamSpec          []*Stream
	name                string
	incomingStreamTypes sets.String
	outgoingStreamType  string
	minInputs           int
	maxInputs           int
	args                []string
	kwargs              KwArgs
	nodeType            string
}

func NewNode(streamSpec []*Stream,
	name string,
	incomingStreamTypes sets.String,
	outgoingStreamType string,
	minInputs int,
	maxInputs int,
	args []string,
	kwargs KwArgs,
	nodeType string) *Node {
	n := &Node{
		streamSpec:          streamSpec,
		name:                name,
		incomingStreamTypes: incomingStreamTypes,
		outgoingStreamType:  outgoingStreamType,
		minInputs:           minInputs,
		maxInputs:           maxInputs,
		args:                args,
		kwargs:              kwargs,
		nodeType:            nodeType,
	}
	n.__checkInputLen()
	n.__checkInputTypes()
	return n
}

func NewInputNode(name string, args []string, kwargs KwArgs) *Node {
	return NewNode(nil,
		name,
		nil,
		"FilterableStream",
		0,
		0,
		args,
		kwargs,
		"InputNode")
}

func NewFilterNode(name string, streamSpec []*Stream, maxInput int, args []string, kwargs KwArgs) *Node {
	return NewNode(streamSpec,
		name,
		sets.NewString("FilterableStream"),
		"FilterableStream",
		1,
		maxInput,
		args,
		kwargs,
		"FilterNode")
}

func NewOutputNode(name string, streamSpec []*Stream, args []string, kwargs KwArgs) *Node {
	return NewNode(streamSpec,
		name,
		sets.NewString("FilterableStream"),
		"OutputStream",
		1,
		-1,
		args,
		kwargs,
		"OutputNode")
}

func NewMergeOutputsNode(name string, streamSpec []*Stream) *Node {
	return NewNode(streamSpec,
		name,
		sets.NewString("OutputStream"),
		"OutputStream",
		1,
		-1,
		nil,
		nil,
		"MergeOutputsNode")
}

func NewGlobalNode(name string, streamSpec []*Stream, args []string, kwargs KwArgs) *Node {
	return NewNode(streamSpec,
		name,
		sets.NewString("OutputStream"),
		"OutputStream",
		1,
		1,
		args,
		kwargs,
		"GlobalNode")
}

func (n *Node) __checkInputLen() {
	streamMap := getStreamMap(n.streamSpec)
	if n.minInputs >= 0 && len(streamMap) < n.minInputs {
		panic(fmt.Sprintf("Expected at least %d input stream(s); got %d", n.minInputs, len(streamMap)))
	}
	if n.maxInputs >= 0 && len(streamMap) > n.maxInputs {
		panic(fmt.Sprintf("Expected at most %d input stream(s); got %d", n.maxInputs, len(streamMap)))
	}
}

func (n *Node) __checkInputTypes() {
	streamMap := getStreamMap(n.streamSpec)
	for _, s := range streamMap {
		if !n.incomingStreamTypes.Has(s.Type) {
			panic(fmt.Sprintf("Expected incoming stream(s) to be of one of the following types: %s; got %s", n.incomingStreamTypes, s.Type))
		}
	}
}

func (n *Node) __getIncomingEdgeMap() map[Label]NodeInfo {
	incomingEdgeMap := map[Label]NodeInfo{}
	streamMap := getStreamMap(n.streamSpec)
	for i, s := range streamMap {
		incomingEdgeMap[Label(fmt.Sprintf("%v", i))] = NodeInfo{
			Node:     s.Node,
			Label:    s.Label,
			Selector: s.Selector,
		}
	}
	return incomingEdgeMap
}

func (n *Node) Hash() int {
	b := 0
	for downStreamLabel, upStreamInfo := range n.IncomingEdgeMap() {
		b += getHash(fmt.Sprintf("%s%d%s%s", downStreamLabel, upStreamInfo.Node.Hash(), upStreamInfo.Label, upStreamInfo.Selector))
	}
	b += getHash(n.args)
	b += getHash(n.kwargs)
	return b
}

func (n *Node) String() string {
	return fmt.Sprintf("%s (%s) <%s>", n.name, getString(n.args), getString(n.kwargs))
}

func (n *Node) Equal(other DagNode) bool {
	return n.Hash() == other.Hash()
}

func (n *Node) ShortRepr() string {
	return n.name
}

func (n *Node) IncomingEdgeMap() map[Label]NodeInfo {
	return n.__getIncomingEdgeMap()
}

func (n *Node) GetInComingEdges() []DagEdge {
	return GetInComingEdges(n, n.IncomingEdgeMap())
}

func (n *Node) Stream(label Label, selector Selector) *Stream {
	return NewStream(n, n.outgoingStreamType, label, selector)
}

func (n *Node) Get(a string) *Stream {
	l := strings.Split(a, ":")
	if len(l) == 2 {
		return n.Stream(Label(l[0]), Selector(l[1]))
	}
	return n.Stream(Label(a), "")
}

func (n *Node) GetFilter(outgoingEdges []DagEdge) string {
	if n.nodeType != "FilterNode" {
		panic("call GetFilter on non-FilterNode")
	}
	args, kwargs, ret := n.args, n.kwargs, ""
	if n.name == "split" || n.name == "asplit" {
		args = []string{fmt.Sprintf("%d", len(outgoingEdges))}
	}
	// args = Args(args).EscapeWith("\\'=:")
	for _, k := range kwargs.EscapeWith("\\'=:").SortedKeys() {
		v := getString(kwargs[k])
		if v != "" {
			args = append(args, fmt.Sprintf("%s=%s", k, v))
		} else {
			args = append(args, fmt.Sprintf("%s", k))
		}
	}
	ret = escapeChars(n.name, "\\'=:")
	if len(args) > 0 {
		ret += fmt.Sprintf("=%s", strings.Join(args, ":"))
	}
	return escapeChars(ret, "\\'[],;")
}
