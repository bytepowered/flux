package zk

import (
	"context"
	"errors"
	"github.com/bytepowered/flux"
	"github.com/bytepowered/flux/logger"
	"github.com/bytepowered/flux/pkg"
	"github.com/bytepowered/flux/remoting"
	"github.com/dubbogo/go-zookeeper/zk"
	"path"
	"strings"
	"sync"
	"time"
)

func NewZookeeperRetriever() *ZookeeperRetriever {
	return &ZookeeperRetriever{
		listenerMap: make(map[string][]remoting.NodeChangedListener),
		quit:        make(chan struct{}),
	}
}

type ZookeeperRetriever struct {
	conn        *zk.Conn
	listenerMap map[string][]remoting.NodeChangedListener
	listenerMu  sync.RWMutex
	quit        chan struct{}
	servers     []string
	timeout     time.Duration
}

// Init 初始化
func (r *ZookeeperRetriever) Init(config *flux.Configuration) error {
	addr := config.GetString("address")
	if "" == addr {
		r.servers = []string{config.GetString("host") + ":" + config.GetString("port")}
	} else {
		r.servers = strings.Split(addr, ",")
	}
	r.timeout = config.GetDuration("timeout")
	return nil
}

// Startup 启动ZK客户端
func (r *ZookeeperRetriever) Startup() error {
	logger.Infow("Zookeeper retriever startup", "server", r.servers)
	conn, _, err := zk.Connect(r.servers, r.timeout, zk.WithLogger(new(zkLogger)))
	if err != nil {
		return err
	}
	r.conn = conn
	return nil
}

// Shutdown 关闭客户端
func (r *ZookeeperRetriever) Shutdown(ctx context.Context) error {
	select {
	case <-r.quit:
		return nil
	default:
		logger.Infow("Zookeeper retriever shutdown", "server", r.servers)
		close(r.quit)
	}
	return nil
}

// Exists 判定指定Path是否存在。注意Path是完整路径。
func (r *ZookeeperRetriever) Exists(path string) (bool, error) {
	b, _, err := r.conn.Exists(path)
	return b, err
}

// Create 创建指定Path的节点
func (r *ZookeeperRetriever) Create(path string) error {
	_, err := r.conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
	return err
}

func (r *ZookeeperRetriever) AddChildrenNodeChangedListener(groupId, parentNodePath string, nodeChangedListener remoting.NodeChangedListener) error {
	if init, err := r.setupListener(groupId, parentNodePath, nodeChangedListener); nil != err {
		return err
	} else if init {
		go r.watchChildrenChanged(parentNodePath)
	}
	return nil
}

// AddNodeChangedListener 添加指定节点的数据变化监听接口
func (r *ZookeeperRetriever) AddNodeChangedListener(groupId, nodePath string, dataChangedListener remoting.NodeChangedListener) error {
	if init, err := r.setupListener(groupId, nodePath, dataChangedListener); nil != err {
		return err
	} else if init {
		go r.watchDataNodeChanged(nodePath)
	}
	return nil
}

func (r *ZookeeperRetriever) watchChildrenChanged(parentNodePath string) {
	logger.Infow("Start watching zk node children", "parent-path", parentNodePath)
	defer func() {
		logger.Infow("Stop watching zk node children, purge listeners", "parent-path", parentNodePath)
		r.listenerMu.Lock()
		delete(r.listenerMap, parentNodePath)
		r.listenerMu.Unlock()
	}()
	handleChildChanged := func(event remoting.NodeEvent) {
		r.listenerMu.RLock()
		listeners, ok := r.listenerMap[parentNodePath]
		r.listenerMu.RUnlock()
		if !ok || 0 == len(listeners) {
			return
		}
		for _, listener := range listeners {
			listener(event)
		}
	}
	cachedChildren := make([]string, 0)
	for {
		newChildren, _, w, err := r.conn.ChildrenW(parentNodePath)
		if nil != err {
			logger.Errorw("Watching zk node children", "parent-path", parentNodePath, "error", err)
			return
		}
		// New: notify
		if len(newChildren) > 0 && len(cachedChildren) == 0 {
			for _, p := range newChildren {
				newChild := path.Join(parentNodePath, p)
				cachedChildren = append(cachedChildren, newChild)
				handleChildChanged(remoting.NodeEvent{
					Path:      newChild,
					EventType: remoting.EventTypeChildAdd,
				})
			}
		}

		select {
		case <-r.quit:
			return

		case zkEvent := <-w.EvtCh:
			logger.Debugw("Receive zk child event", "event", zkEvent)
			if zkEvent.Type == zk.EventNodeChildrenChanged {
				newChildren, _, err := r.conn.Children(zkEvent.Path)
				if nil != err {
					logger.Errorw("get children data", "parent-path", parentNodePath, "error", err)
					return
				}
				// Add
				for i, p := range newChildren {
					newChildren[i] = path.Join(parentNodePath, p) // Update full path
					if !pkg.StringSliceContains(cachedChildren, newChildren[i]) {
						handleChildChanged(remoting.NodeEvent{
							Path:      newChildren[i],
							EventType: remoting.EventTypeChildAdd,
						})
					}
				}
				// Deleted
				for _, p := range cachedChildren {
					if !pkg.StringSliceContains(newChildren, p) {
						handleChildChanged(remoting.NodeEvent{
							Path:      p,
							EventType: remoting.EventTypeChildDelete,
						})
					}
				}
				cachedChildren = newChildren
			}
		}
	}
}

func (r *ZookeeperRetriever) watchDataNodeChanged(nodePath string) {
	logger.Infow("Start watching zk node data", "node-path", nodePath)
	defer func() {
		logger.Infow("Stop watching zk node data, purge listeners", "node-path", nodePath)
		r.listenerMu.Lock()
		delete(r.listenerMap, nodePath)
		r.listenerMu.Unlock()
	}()
	inited := false
	for {
		_, _, w, err := r.conn.ExistsW(nodePath)
		if nil != err {
			logger.Errorw("Watching zk node data", "node-path", nodePath, "error", err)
			return
		}
		if !inited {
			w.EvtCh <- zk.Event{
				Type:  zk.EventNodeCreated,
				State: zk.StateUnknown,
				Path:  nodePath,
			}
			inited = true
		}
		select {
		case <-r.quit:
			return

		case zkEvent := <-w.EvtCh:
			logger.Debugw("Receive zk data event", "event", zkEvent)
			var (
				eventType remoting.EventType
				eventData []byte
			)
			r.listenerMu.RLock()
			listeners, ok := r.listenerMap[zkEvent.Path]
			r.listenerMu.RUnlock()
			if !ok || 0 == len(listeners) {
				continue
			}
			const msgErrGetNodeData = "zk get node data"
			switch zkEvent.Type {
			case zk.EventNodeDataChanged:
				data, _, err := r.conn.Get(zkEvent.Path)
				if nil != err {
					logger.Errorw(msgErrGetNodeData, "node-path", nodePath, "error", err)
					return
				}
				eventType = remoting.EventTypeNodeUpdate
				eventData = data
			case zk.EventNodeCreated:
				data, _, err := r.conn.Get(zkEvent.Path)
				if nil != err {
					logger.Errorf(msgErrGetNodeData, "node-path", nodePath, "error", err)
					return
				}
				eventType = remoting.EventTypeNodeAdd
				eventData = data
			case zk.EventNodeDeleted:
				eventType = remoting.EventTypeNodeDelete
			default:
				continue
			}
			event := remoting.NodeEvent{
				Path:      zkEvent.Path,
				EventType: eventType,
				Data:      eventData,
			}
			for _, listener := range listeners {
				listener(event)
			}
		}
	}
}

func (r *ZookeeperRetriever) setupListener(groupId, nodeKey string, listener remoting.NodeChangedListener) (bool, error) {
	if groupId != "" {
		logger.Warnw("Zookeeper not support groupId", "groupId", groupId)
	}
	if nodeKey == "" {
		return false, errors.New("invalid node key: empty")
	}
	if nil == listener {
		return false, errors.New("invalid listener: nil")
	}
	r.listenerMu.Lock()
	defer r.listenerMu.Unlock()
	if ls, ok := r.listenerMap[nodeKey]; ok {
		r.listenerMap[nodeKey] = append(ls, listener)
		return false, nil
	} else {
		r.listenerMap[nodeKey] = []remoting.NodeChangedListener{listener}
		return true, nil
	}
}

////

type zkLogger int

func (zkLogger) Printf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}
