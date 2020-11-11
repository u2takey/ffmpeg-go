// +build !debug

package ffmpeg_go

func DebugNodes(node []DagNode) {}

func DebugOutGoingMap(node []DagNode, m map[int]map[Label][]NodeInfo) {}
