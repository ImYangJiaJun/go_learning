package pipeline

import "sync"

// gen 把 nums 逐个发送到新建的 channel 上并返回它。
// 发送在内部 goroutine 中进行：发完所有数后关闭 out；
// 中途 done 关闭则立即停止发送并关闭 out（谁生产谁关闭）。
func gen(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case <-done:
				return
			case out <- n:
			}
		}
	}()
	return out
}

// sq 从 in 读整数，把平方值发送到新建的 channel 返回。
// in 关闭（自然结束）或 done 关闭（被取消）时，退出并关闭 out。
func sq(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case <-done:
				return
			case out <- n * n:
			}
		}
	}()
	return out
}

// merge 把 cs 里任意多个 channel 的数据搬到同一个 channel 返回（扇入）。
// 每个输入 channel 配一个搬运 goroutine；全部搬完（或 done 关闭）后关闭 out。
func merge(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)
	for _, c := range cs {
		wg.Add(1)
		go func(in <-chan int) {
			defer wg.Done()
			for n := range in {
				select {
				case <-done:
					return
				case out <- n:
				}
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// workerPool 启动 n 个 worker goroutine 共同从 jobs 读任务，
// 把每个任务的平方值发送到新建的 channel 返回。
// jobs 关闭且全部结果发完（或 done 关闭）后关闭 out。
// 注意：out 只能由 workerPool 在"所有 worker 都退出后"统一关闭，
// 任何单个 worker 都不得 close(out)——其他 worker 还在往上面发。
func workerPool(done <-chan struct{}, jobs <-chan int, n int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)
	for range n {
		wg.Go(func() {
			for j := range jobs {
				select {
				case <-done:
					return
				case out <- j * j:
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
