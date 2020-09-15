package main

import (
	"fmt"
	"github.com/real3stone/geecache/geecache"
	"log"
	"net/http"
	"testing"
)

// 构造访问请求： curl http://localhost:9999/_geecache/scores/Tom
func TestHTTPPoolGet(t *testing.T) {
	// 创建一个Cache空间
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		// 定义回调函数
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"   // 定义服务地址并传入
	peers := geecache.NewHTTPPool(addr)
	http.ListenAndServe(addr, peers)
	//log.Println("geecache is running at", addr)
	//log.Fatal(http.ListenAndServe(addr, peers))
}

