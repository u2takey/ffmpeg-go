package ffmpeg_go

import (
	"bytes"
	"fmt"
)

type ViewType string

const (
	// FlowChart the diagram type for output in flowchart style (https://mermaid-js.github.io/mermaid/#/flowchart) (including current state
	ViewTypeFlowChart ViewType = "flowChart"
	// StateDiagram the diagram type for output in stateDiagram style (https://mermaid-js.github.io/mermaid/#/stateDiagram)
	ViewTypeStateDiagram ViewType = "stateDiagram"
)

func (s *Stream) View(viewType ViewType) (string, error) {
	switch viewType {
	case ViewTypeFlowChart:
		return visualizeForMermaidAsFlowChart(s)
	case ViewTypeStateDiagram:
		return visualizeForMermaidAsStateDiagram(s)
	default:
		return "", fmt.Errorf("unknown ViewType: %s", viewType)
	}
}

func visualizeForMermaidAsStateDiagram(s *Stream) (string, error) {
	var buf bytes.Buffer

	nodes := getStreamSpecNodes([]*Stream{s})
	var dagNodes []DagNode
	for i := range nodes {
		dagNodes = append(dagNodes, nodes[i])
	}
	sorted, outGoingMap, err := TopSort(dagNodes)
	if err != nil {
		return "", err
	}

	buf.WriteString("stateDiagram\n")

	for _, node := range sorted {
		next := outGoingMap[node.Hash()]
		for k, v := range next {
			for _, nextNode := range v {
				label := string(k)
				if label == "" {
					label = "<>"
				}
				buf.WriteString(fmt.Sprintf(`    %s --> %s: %s`, node.ShortRepr(), nextNode.Node.ShortRepr(), label))
				buf.WriteString("\n")
			}
		}
	}
	return buf.String(), nil
}

// visualizeForMermaidAsFlowChart outputs a visualization of a FSM in Mermaid format (including highlighting of current state).
func visualizeForMermaidAsFlowChart(s *Stream) (string, error) {
	var buf bytes.Buffer

	nodes := getStreamSpecNodes([]*Stream{s})
	var dagNodes []DagNode
	for i := range nodes {
		dagNodes = append(dagNodes, nodes[i])
	}
	sorted, outGoingMap, err := TopSort(dagNodes)
	if err != nil {
		return "", err
	}
	buf.WriteString("graph LR\n")

	for _, node := range sorted {
		buf.WriteString(fmt.Sprintf(`    %d[%s]`, node.Hash(), node.ShortRepr()))
		buf.WriteString("\n")
	}
	buf.WriteString("\n")

	for _, node := range sorted {
		next := outGoingMap[node.Hash()]
		for k, v := range next {
			for _, nextNode := range v {
				// todo ignore merged output
				label := string(k)
				if label == "" {
					label = "<>"
				}
				buf.WriteString(fmt.Sprintf(`    %d --> |%s| %d`, node.Hash(), fmt.Sprintf("%s:%s", nextNode.Label, label), nextNode.Node.Hash()))
				buf.WriteString("\n")
			}
		}
	}

	buf.WriteString("\n")

	return buf.String(), nil
}
