package main

import (
	"fmt"
	"github.com/real3stone/geecache/geecache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
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
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
