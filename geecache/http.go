package geecache

import (
	"fmt"
	"github.com/real3stone/geecache/geecache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/*
	结构体 HTTPPool，作为承载节点间 HTTP 通信的核心数据结构（包括服务端和客户端)
*/
const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)


// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self     string // 记录自己的地址，包括主机名/IP 和端口  e.g. "https://example.net:8000"
	basePath string // 节点间通讯地址的前缀，默认是 /_geecache/

	// 为 HTTPPool 添加节点选择的功能
	mu sync.Mutex   // guards peers and httpGetters
	peers *consistenthash.Map // 新增成员变量 peers，类型是一致性哈希算法的 Map，用来根据具体的 key 选择节点
	httpGetters map[string]*httpGetter // 映射远程节点与对应的 httpGetter
	// 每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关
}

// initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool{
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}

// ------------------------------------------------------------
// 记录日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Println("[Server %s] %s", p.self, fmt.Sprintf(format, v...)) // 格式化输出
}

/*
	核心方法，实现ServeHTTP 方法
*/
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2) //从basePath之后的子串开始切分，得到<groupname>/<key>
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)   //得到缓存空间
	if group == nil {
		http.Error(w, "no such group: "+ groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)  // 获取缓存
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())    // 把获得的数据写入Response
}


// ---------------------------------------------------------------------------
/*
	客户端
*/
type httpGetter struct {
	baseURL string // 将要访问的远程节点的地址，例如 http://example.com/_geecache/
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(  //
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// 请求成功
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

//var _ PeerGetter = (*httpGetter)(nil)

/*
	实现PickPeer接口
*/
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)   //实例化一致性哈希算法
	p.peers.Add(peers...)    // 添加传入的节点
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers { // 为每个节点创建一个HTTP客户端httpGetter
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
// PickerPeer() 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		//return p.httpGetters[peer], true  // 一直报错，还想哪里粘得有问题？
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)


// 至此，HTTPPool 既具备了提供 HTTP 服务的能力，也具备了根据具体的 key，创建 HTTP 客户端从远程节点获取缓存值的能力
