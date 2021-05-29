package remoting

import "fmt"

type NoteEventType int

const (
	EventTypeNodeAdd = iota
	EventTypeNodeDelete
	EventTypeNodeUpdate
	EventTypeChildAdd
	EventTypeChildDelete
	eventTypeUndefined
)

var nodeEventNames = [...]string{
	"NodeAdd",
	"NodeDelete",
	"NodeUpdate",
	"ChildAdd",
	"ChildDelete",
	"Undefined",
}

func (t NoteEventType) String() string {
	if t > eventTypeUndefined {
		return nodeEventNames[eventTypeUndefined]
	}
	return nodeEventNames[t]
}

type NodeEvent struct {
	Path  string
	Event NoteEventType
	Data  []byte
}

func (e NodeEvent) String() string {
	return fmt.Sprintf("Event{Type{%s}, Path{%s} Body{%s}}", e.Event, e.Path, string(e.Data))
}

type NodeChangedListener func(event NodeEvent)

type NodeRetriever interface {
	// AddChangedListener 监听数据节点的数据变更
	AddChangedListener(groupId, dataNodeKey string, dataListener NodeChangedListener) error

	// AddChildrenNodeChangedListener 监听目录节点的子节点变更
	AddChildChangedListener(groupId, dirNodeKey string, childListener NodeChangedListener) error
}
