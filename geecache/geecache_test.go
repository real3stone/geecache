package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

// 测试 保证数据获取的回调函数能够正常工作
func TestGetter(t *testing.T) {
	// 借助 GetterFunc 的类型转换，将一个匿名回调函数转换成了接口 f Getter
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

// 定义map，模拟数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 测试Group的Get方法
func TestGroupGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db)) // 记录查询数据库的次数
	// 创建一个缓存空间
	gee := NewGroup("scores", 2<<10, GetterFunc(
		// 自定义回调函数 传入
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key：", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok { //数据库的初次查询
					loadCounts[key] = 0
				}
				loadCounts[key] += 1 //累加查询数据库的次数
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)  // 数据库中也没有获取到
		}))

	// 依次查询一下DB中的三个值
	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v { //先取一次，然后缓存中就有了
			t.Fatal("failed to get value of Tom")
		} // load from callback function
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}

}


