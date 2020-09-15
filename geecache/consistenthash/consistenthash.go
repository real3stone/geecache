package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 函数类型 Hash，采取依赖注入的方式
type Hash func(data []byte) uint32

// 一致性哈希算法的主数据结构
type Map struct {
	hash     Hash           // Hash 函数
	replicas int            // 虚拟节点倍数 （每一个真实节点，对应创建 m.replicas 个虚拟节点）
	keys     []int          // Sorted 哈希环（存储所有虚拟节点，并排序）
	hashMap  map[int]string // 虚拟节点与真实节点的映射表 (键是虚拟节点的哈希值，值是真实节点的名称)
}


// 构造函数 New() 允许自定义虚拟节点倍数和 Hash 函数
func New(replicas int, fn Hash) *Map {
	m := &Map {
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil { //没有传入自定义hash算法，就用库中的
		m.hash = crc32.ChecksumIEEE
	}
	return m
}


// 添加真实节点
// Add： adds some keys to the hash.
func (m *Map) Add(keys ...string) { // 允许传入 0 或 多个真实节点的名称
	for _, key := range keys {      // 对每一个真实节点 key，对应创建 m.replicas 个虚拟节点
		for i := 0; i < m.replicas; i++ { //虚拟节点的名称是：strconv.Itoa(i) + key，即通过添加编号的方式区分不同虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash) //加入哈希环中
			m.hashMap[hash] = key         //添加映射关系（虚拟节点到真实节点）
		}
	}

	// Ints：sorts a slice of ints in increasing order
	sort.Ints(m.keys) // 环上虚拟节点key排序
}

// Get gets the closest item in the hash to the provided key.
/*
第一步，计算 key 的哈希值。
第二步，顺时针找到第一个匹配的虚拟节点的下标 idx，从 m.keys 中获取到对应的哈希值。
如果 idx == len(m.keys)，说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
第三步，通过 hashMap 映射得到真实的节点。
*/
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key))) // 计算key的哈希值
	// Binary search for appropriate replica. 二分查找虚拟节点 (传入的是左右切分函数)
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]] // 环形，取余
}
