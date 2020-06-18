package consistentHash

import (
	"fmt"
	"hash/crc32"
	"sort"
)

const defaultReplicas = 5

type hash func([]byte) uint32

type Hash struct {
	h        hash // hash 函数
	replicas int  // 副本数
	set      map[int]string
	keys     []int
}

func NewHash(replice int, h hash) *Hash {
	if h == nil {
		h = crc32.ChecksumIEEE
	}
	if replice <= 0 {
		replice = defaultReplicas
	}
	return &Hash{
		h:        h,
		replicas: replice,
		set:      make(map[int]string),
		keys:     make([]int, 0),
	}
}

func (h *Hash) Add(v string) {
	for i := 0; i < h.replicas; i++ {
		key := int(h.h([]byte(fmt.Sprintf("%s-%d", v, i))))
		h.keys = append(h.keys, key)
		h.set[key] = v
	}
	sort.Ints(h.keys)
}

func (h *Hash) Get(v string) string {
	key := int(h.h([]byte(v)))

	l := len(h.keys)

	index := sort.Search(l, func(i int) bool {
		return h.keys[i] >= key
	})
	return h.set[h.keys[(index%l)]]
}
