package core

import (
	"context"
	"sync"
)

// Collect 把 jobs 交给固定数量的 worker 处理，并收集全部结果。
// 三条退出路径都写明了：发送方发完后关闭 jobs；worker 见 jobs 关闭或 ctx 取消后退出；
// 收集者见 out 关闭后退出。WaitGroup 负责等所有人收敛，避免 goroutine 泄漏。
func Collect(ctx context.Context, jobs []int, workers int) ([]int, error) {
	if workers < 1 {
		workers = 1
	}
	jobCh := make(chan int)
	outCh := make(chan int)
	var wg sync.WaitGroup
	var collectWg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobCh:
					if !ok {
						return
					}
					outCh <- j * 2
				}
			}
		}()
	}

	results := make([]int, 0, len(jobs))
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for v := range outCh {
			results = append(results, v)
		}
	}()

	// 关闭规则：只由发送方关闭 jobs，接收方永不关闭。
	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			// 提前取消检查：ctx 已取消时确定不再投递任何任务，
			// 使"取消后零结果"成为可判定契约（删掉它则红—绿对照会变红）。
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case jobCh <- j:
			}
		}
	}()

	wg.Wait()
	close(outCh) // 所有 worker 退出后，由协调方关闭结果通道
	collectWg.Wait()
	return results, ctx.Err()
}
