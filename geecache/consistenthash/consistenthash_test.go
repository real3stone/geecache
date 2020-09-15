package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 自定义hash算法 （只是进行一下简单的字符串到数字的转换）；虚拟节点倍数为3
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2") //有 2/4/6 三个真实节点
	// 对应上述真实节点的虚拟节点的哈希值是 02/12/22、04/14/24、06/16/26

	// 测试：
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// Adds 8, 18, 28
	hash.Add("8")

	// 27 should now map to 8.
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

}
