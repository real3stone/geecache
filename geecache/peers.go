package geecache


// PeerPicker is the interface that must be implemented to locate the peer that owns a specific key.
type PeerPicker interface {
	// 根据传入的 key 选择相应节点 PeerGetter
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
// PeerGetter对应HTTP客户端，在http.go中进行创建
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

