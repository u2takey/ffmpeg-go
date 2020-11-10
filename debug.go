// +build debug

package ffmpeg_go

import (
	"fmt"
	"log"
	"strings"
)

func DebugNodes(node []DagNode) {
	b := strings.Builder{}
	for _, n := range node {
		b.WriteString(fmt.Sprintf("%s\n", n.String()))
	}
	log.Println(b.String())
}

func DebugOutGoingMap(node []DagNode, m map[int]map[Label][]NodeInfo) {
	b := strings.Builder{}
	h := map[int]DagNode{}
	for _, n := range node {
		h[n.Hash()] = n
	}
	for k, v := range m {
		b.WriteString(fmt.Sprintf("[Key]: %s", h[k].String()))
		b.WriteString(" [Value]: {")
		for l, mm := range v {
			if l == "" {
				l = "None"
			}
			b.WriteString(fmt.Sprintf("%s: [", l))
			for _, x := range mm {
				b.WriteString(x.Node.String())
				b.WriteString(", ")
			}
			b.WriteString("]")
		}
		b.WriteString("}")
		b.WriteString("\n")
	}
	log.Println(b.String())
}
