package zookeeper

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

func NewZkRetriever() *ZkRetriever {
	return &ZkRetriever{
		listenerMap: make(map[string][]remoting.NodeChangedListener),
		quit:        make(chan struct{}),
	}
}

type ZkRetriever struct {
	conn        *zk.Conn
	listenerMap map[string][]remoting.NodeChangedListener
	listenerMu  sync.RWMutex
	quit        chan struct{}
	servers     []string
	timeout     time.Duration
}

func (r *ZkRetriever) InitWith(config *flux.Configuration) error {
	addr := config.GetString("address")
	if "" == addr {
		r.servers = []string{config.GetString("host") + ":" + config.GetString("port")}
	} else {
		r.servers = strings.Split(addr, ",")
	}
	r.timeout = config.GetDuration("timeout")
	return nil
}

func (r *ZkRetriever) Startup() error {
	logger.Infow("ZkRetriver startup", "server", r.servers)
	conn, _, err := zk.Connect(r.servers, r.timeout, zk.WithLogger(new(zkLogger)))
	if err != nil {
		return err
	}
	r.conn = conn
	return nil
}

func (r *ZkRetriever) Shutdown(ctx context.Context) error {
	select {
	case <-r.quit:
		return nil
	default:
		logger.Infow("ZkRetriver shutdown", "server", r.servers)
		close(r.quit)
	}
	return nil
}

func (r *ZkRetriever) Exists(path string) (bool, error) {
	b, _, err := r.conn.Exists(path)
	return b, err
}

func (r *ZkRetriever) Create(path string) error {
	_, err := r.conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
	return err
}

func (r *ZkRetriever) WatchChildren(groupId, dirKey string, childChangedListener remoting.NodeChangedListener) error {
	if init, err := r.setupListener(groupId, dirKey, childChangedListener); nil != err {
		return err
	} else if init {
		go r.watchChildrenChanged(dirKey)
	}
	return nil
}

func (r *ZkRetriever) WatchNodeData(groupId, nodeKey string, dataChangedListener remoting.NodeChangedListener) error {
	if init, err := r.setupListener(groupId, nodeKey, dataChangedListener); nil != err {
		return err
	} else if init {
		go r.watchDataNodeChanged(nodeKey)
	}
	return nil
}

func (r *ZkRetriever) setupListener(groupId, nodeKey string, listener remoting.NodeChangedListener) (bool, error) {
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

func (r *ZkRetriever) watchChildrenChanged(dirKey string) {
	logger.Infow("Start watching zk node children", "path", dirKey)
	defer func() {
		logger.Errorf("Stop watching zk node children, purge listeners", "path", dirKey)
		r.listenerMu.Lock()
		delete(r.listenerMap, dirKey)
		r.listenerMu.Unlock()
	}()
	handleChildChanged := func(event remoting.NodeEvent) {
		r.listenerMu.RLock()
		listeners, ok := r.listenerMap[dirKey]
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
		newChildren, _, w, err := r.conn.ChildrenW(dirKey)
		if nil != err {
			logger.Errorw("Watching zk node children,", "path", dirKey, "error", err)
			return
		}
		// New: notify
		if len(newChildren) > 0 && len(cachedChildren) == 0 {
			for _, p := range newChildren {
				newChild := path.Join(dirKey, p)
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
			logger.Debugf("Receive zk child event{type:%s, server:%s, path:%s, state:%d-%s, err:%s}",
				zkEvent.Type, zkEvent.Server, zkEvent.Path, zkEvent.State, zkEvent.State, zkEvent.Err)
			if zkEvent.Type == zk.EventNodeChildrenChanged {
				newChildren, _, err := r.conn.Children(zkEvent.Path)
				if nil != err {
					logger.Errorw("get children data", "path", dirKey, "error", err)
					return
				}
				// Add
				for i, p := range newChildren {
					newChildren[i] = path.Join(dirKey, p) // Update full path
					if !pkg.StrContains(newChildren[i], cachedChildren) {
						handleChildChanged(remoting.NodeEvent{
							Path:      newChildren[i],
							EventType: remoting.EventTypeChildAdd,
						})
					}
				}
				// Deleted
				for _, p := range cachedChildren {
					if !pkg.StrContains(p, newChildren) {
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

func (r *ZkRetriever) watchDataNodeChanged(nodePath string) {
	logger.Infof("Start watching zk node data(%s)", nodePath)
	defer func() {
		logger.Errorf("Stop watching zk node data(%s), purge listeners", nodePath)
		r.listenerMu.Lock()
		delete(r.listenerMap, nodePath)
		r.listenerMu.Unlock()
	}()
	inited := false
	for {
		_, _, w, err := r.conn.ExistsW(nodePath)
		if nil != err {
			logger.Errorf("Watching zk node data(%s), err: %s", nodePath, err)
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
			logger.Debugf("Receive zk data event{type:%s, server:%s, path:%s, state:%d-%s, err:%s}",
				zkEvent.Type, zkEvent.Server, zkEvent.Path, zkEvent.State, zkEvent.State, zkEvent.Err)
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
			const msgErrGetNodeData = "zk get node data, path:%s, err: %s"
			switch zkEvent.Type {
			case zk.EventNodeDataChanged:
				data, _, err := r.conn.Get(zkEvent.Path)
				if nil != err {
					logger.Errorf(msgErrGetNodeData, nodePath, err)
					return
				}
				eventType = remoting.EventTypeNodeUpdate
				eventData = data
			case zk.EventNodeCreated:
				data, _, err := r.conn.Get(zkEvent.Path)
				if nil != err {
					logger.Errorf(msgErrGetNodeData, nodePath, err)
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

////

type zkLogger int

func (zkLogger) Printf(format string, a ...interface{}) {
	logger.Debugf(format, a...)
}
