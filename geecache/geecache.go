package geecache

import (
	"fmt"
	"log"
	"sync"
)

/*
	设计了一个回调函数(callback)，在缓存不存在时，调用这个函数，得到源数据
	ps：如何从源头获取数据，应该是用户决定的事情，我们就把这件事交给用户好了
*/
type Getter interface {
	Get(key string) ([]byte, error)
}

// 函数类型 GetterFunc, 实现Getter接口的Get方法
type GetterFunc func(key string) ([]byte, error)

// 之后就能把自定义的获取源数据的函数直接作为参数传进来
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// -----------------------
//  一个 Group 可以认为是一个缓存的命名空间
type Group struct {
	name      string // 一个 Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 name
	getter    Getter // 回调(callback): 缓存未命中时获取源数据
	mainCache cache  // 封装的lru，可并发的缓存

	peers	  PeerPicker    // 分布式获取缓存
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)  // 所有缓存空间
)

// 新建一个缓存空间
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}

	groups[name] = g  //将 group 存储在全局变量 groups 中
	return g
}


// 获取某个缓存空间
func GetGroup(name string) *Group{
	mu.RLock()  // 不涉及任何冲突变量的写操作,所以用只读锁即可
	g := groups[name]
	mu.RUnlock()
	return g
}

// 获取缓存
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("[GeeCache] hit!")
		return v, nil
	}

	return g.load(key)    // 缓存失效，从DB获取数据
}

/*
	load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），
	getLocally() 调用用户回调函数 g.getter.Get() 获取源数据，
	并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
*/
func (g *Group) load(key string) (ByteView, error) {
	if g.peers != nil {
		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err := g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", ok)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// 单机场景下获取数据（分布式场景下用另外api）
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}  // 获取数据副本
	g.populateCache(key, value)  // 将数据填入Cache空间中
	return value, nil
}

// 没明白作者这儿还要单独再封装一下，后面要复用？
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}


/*
	Group的方法： RegisterPeers() ，将 实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
*/
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

