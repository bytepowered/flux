package remoting

import "fmt"

type EventType int

const (
	EventTypeNodeAdd = iota
	EventTypeNodeDelete
	EventTypeNodeUpdate
	EventTypeChildAdd
	EventTypeChildDelete
)

var _eventTypeNames = [...]string{
	"nodeAdd",
	"nodeDelete",
	"nodeUpdate",
	"childAdd",
	"childDelete",
}

func (t EventType) String() string {
	return _eventTypeNames[t]
}

type NodeEvent struct {
	Path      string
	EventType EventType
	Data      []byte
}

func (e NodeEvent) String() string {
	return fmt.Sprintf("Event{Type{%s}, Path{%s} Data{%s}}", e.EventType, e.Path, string(e.Data))
}

type NodeChangedListener func(event NodeEvent)

type NodeRetriever interface {
	// WatchNodeData 监听数据节点的数据变更
	WatchNodeData(groupId, dataNodeKey string, dataListener NodeChangedListener) error

	// WatchChildren 监听目录节点的子节点变更
	WatchChildren(groupId, dirNodeKey string, childrenListener NodeChangedListener) error
}
