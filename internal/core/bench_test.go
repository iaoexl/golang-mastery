package core

import "testing"

// benchSink 保证基准函数的计算结果被真正消费，避免整段逻辑被优化掉。
var benchSink int

func BenchmarkParseSpec(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSink, _ = ParseSpec("region-a:42")
	}
}

// FuzzParseSpec 的种子语料覆盖合法、空串、负数与越界四类，
// 不变量是"只要返回 nil 错误，结果就必须落在约定区间内"。
func FuzzParseSpec(f *testing.F) {
	f.Add("region-a:42")
	f.Add("")
	f.Add("x:-1")
	f.Add("x:1001")
	f.Fuzz(func(t *testing.T, s string) {
		n, err := ParseSpec(s)
		if err == nil && (n < 0 || n > 1000) {
			t.Fatalf("ParseSpec(%q) = %d, out of contract", s, n)
		}
	})
}
