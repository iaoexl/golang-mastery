package core

// Snapshot 返回不共享底层数组的独立副本。
// 契约写在函数名上：调用方拿到的是快照，改它不会影响原数据。
func Snapshot(s []int) []int {
	out := make([]int, len(s))
	copy(out, s)
	return out
}

// Window 用完整切片表达式把容量限制到 hi，返回一个只读窗口。
// 因为 cap 被钉死，对返回值 append 必然触发新分配，不会悄悄改写原数组。
func Window(s []int, lo, hi int) []int {
	return s[lo:hi:hi]
}
