package core

import "sync"

// Storer 由调用方定义的最小接口，实现方不需要 import 调用方包。
type Storer interface {
	Put(key string, n int) error
	Get(key string) (int, bool)
}

// MemoryStore 用零值不可直接写 map 这一事实说明：map 需要显式 make。
type MemoryStore struct {
	mu sync.Mutex
	m  map[string]int
}

// NewMemoryStore 构造一个可用实例。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{m: make(map[string]int)}
}

// Put 写入一条计数，临界区只包住 map 操作本身。
func (s *MemoryStore) Put(key string, n int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = n
	return nil
}

// Get 读取一条计数。
func (s *MemoryStore) Get(key string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.m[key]
	return v, ok
}

// 编译期接口断言：实现方变了方法集，这里立刻报错，不必等运行时。
var _ Storer = (*MemoryStore)(nil)

// Number 是类型集合约束：底层类型属于这三族之一的具名类型都可实例化。
// 它与"方法集"是两条不同的判定链，不要混为一谈。
type Number interface {
	~int | ~int64 | ~float64
}

// Sum 对任意数值族切片求和，算法相同而类型不同，正是泛型的适用场景。
func Sum[N Number](xs []N) N {
	var total N
	for _, x := range xs {
		total += x
	}
	return total
}
