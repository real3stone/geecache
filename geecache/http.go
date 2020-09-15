package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

/*
	结构体 HTTPPool，作为承载节点间 HTTP 通信的核心数据结构（包括服务端和客户端)
*/
const defaultBasePath = "/_geecache/"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	self     string // 记录自己的地址，包括主机名/IP 和端口  e.g. "https://example.net:8000"
	basePath string // 节点间通讯地址的前缀，默认是 /_geecache/
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
