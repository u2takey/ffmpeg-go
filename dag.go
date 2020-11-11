package ffmpeg_go

import (
	"errors"
)

// Node in a directed-acyclic graph (DAG).
//
// Edges:
// DagNodes are connected by edges.  An edge connects two nodes with a label for each side:
// - ``upstream_node``: upstream/parent node
// - ``upstream_label``: label on the outgoing side of the upstream node
// - ``downstream_node``: downstream/child node
// - ``downstream_label``: label on the incoming side of the downstream node
//
// For example, DagNode A may be connected to DagNode B with an edge labelled "foo" on A's side, and "bar" on B's
// side:
//
// _____               _____
// |     |             |     |
// |  A  >[foo]---[bar]>  B  |
// |_____|             |_____|
//
// Edge labels may be integers or strings, and nodes cannot have more than one incoming edge with the same label.
//
// DagNodes may have any number of incoming edges and any number of outgoing edges.  DagNodes keep track only of
// their incoming edges, but the entire graph structure can be inferred by looking at the furthest downstream
// nodes and working backwards.
//
// Hashing:
// DagNodes must be hashable, and two nodes are considered to be equivalent if they have the same hash value.
//
// Nodes are immutable, and the hash should remain constant as a result.  If a node with new contents is required,
// create a new node and throw the old one away.
//
// String representation:
// In order for graph visualization tools to show useful information, nodes must be representable as strings.  The
// ``String`` operator should provide a more or less "full" representation of the node, and the ``ShortRepr``
// property should be a shortened, concise representation.
//
// Again, because nodes are immutable, the string representations should remain constant.
type DagNode interface {
	Hash() int
	// Compare two nodes
	Equal(other DagNode) bool
	// Return a full string representation of the node.
	String() string
	// Return a partial/concise representation of the node
	ShortRepr() string
	// Provides information about all incoming edges that connect to this node.
	//
	//        The edge map is a dictionary that maps an ``incoming_label`` to ``(outgoing_node, outgoing_label)``.  Note that
	//        implicity, ``incoming_node`` is ``self``.  See "Edges" section above.
	IncomingEdgeMap() map[Label]NodeInfo
}

type Label string
type NodeInfo struct {
	Node     DagNode
	Label    Label
	Selector Selector
}
type Selector string

type DagEdge struct {
	DownStreamNode   DagNode
	DownStreamLabel  Label
	UpStreamNode     DagNode
	UpStreamLabel    Label
	UpStreamSelector Selector
}

func GetInComingEdges(downStreamNode DagNode, inComingEdgeMap map[Label]NodeInfo) []DagEdge {
	var edges []DagEdge
	for _, downStreamLabel := range _getAllLabelsInSorted(inComingEdgeMap) {
		upStreamInfo := inComingEdgeMap[downStreamLabel]
		edges = append(edges, DagEdge{
			DownStreamNode:   downStreamNode,
			DownStreamLabel:  downStreamLabel,
			UpStreamNode:     upStreamInfo.Node,
			UpStreamLabel:    upStreamInfo.Label,
			UpStreamSelector: upStreamInfo.Selector,
		})
	}
	return edges
}

func GetOutGoingEdges(upStreamNode DagNode, outOutingEdgeMap map[Label][]NodeInfo) []DagEdge {
	var edges []DagEdge
	for _, upStreamLabel := range _getAllLabelsSorted(outOutingEdgeMap) {
		downStreamInfos := outOutingEdgeMap[upStreamLabel]
		for _, downStreamInfo := range downStreamInfos {
			edges = append(edges, DagEdge{
				DownStreamNode:   downStreamInfo.Node,
				DownStreamLabel:  downStreamInfo.Label,
				UpStreamNode:     upStreamNode,
				UpStreamLabel:    upStreamLabel,
				UpStreamSelector: downStreamInfo.Selector,
			})
		}

	}
	return edges
}

func TopSort(downStreamNodes []DagNode) (sortedNodes []DagNode, outOutingEdgeMaps map[int]map[Label][]NodeInfo, err error) {
	markedNodes := map[int]struct{}{}
	markedSortedNodes := map[int]struct{}{}
	outOutingEdgeMaps = map[int]map[Label][]NodeInfo{}

	var visit func(upStreamNode, downstreamNode DagNode, upStreamLabel, downStreamLabel Label, downStreamSelector Selector) error
	visit = func(upStreamNode, downstreamNode DagNode, upStreamLabel, downStreamLabel Label, downStreamSelector Selector) error {
		if _, ok := markedNodes[upStreamNode.Hash()]; ok {
			return errors.New("graph if not DAG")
		}
		if downstreamNode != nil {
			if a, ok := outOutingEdgeMaps[upStreamNode.Hash()]; !ok || a == nil {
				outOutingEdgeMaps[upStreamNode.Hash()] = map[Label][]NodeInfo{}
			}
			outgoingEdgeMap := outOutingEdgeMaps[upStreamNode.Hash()]
			outgoingEdgeMap[upStreamLabel] = append(outgoingEdgeMap[upStreamLabel], NodeInfo{
				Node:     downstreamNode,
				Label:    downStreamLabel,
				Selector: downStreamSelector,
			})
		}

		if _, ok := markedSortedNodes[upStreamNode.Hash()]; !ok {
			markedNodes[upStreamNode.Hash()] = struct{}{}
			for _, edge := range GetInComingEdges(upStreamNode, upStreamNode.IncomingEdgeMap()) {
				err := visit(edge.UpStreamNode, edge.DownStreamNode, edge.UpStreamLabel, edge.DownStreamLabel, edge.UpStreamSelector)
				if err != nil {
					return err
				}
			}
			delete(markedNodes, upStreamNode.Hash())
			sortedNodes = append(sortedNodes, upStreamNode)
			markedSortedNodes[upStreamNode.Hash()] = struct{}{}
		}
		return nil
	}

	for len(downStreamNodes) > 0 {
		node := downStreamNodes[len(downStreamNodes)-1]
		downStreamNodes = downStreamNodes[:len(downStreamNodes)-1]
		err = visit(node, nil, "", "", "")
		if err != nil {
			return
		}
	}
	return
}
