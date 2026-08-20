package pipeline

import (
	"slices"
	"testing"
	"time"
)

func TestPipeline(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	in := gen(done, 2, 3) // 一个数据源
	out1 := sq(done, in)  // 两个 sq 从同一个 in 抢数据
	out2 := sq(done, in)
	result := merge(done, out1, out2)
	res := make([]int, 0)
	for v := range result {
		res = append(res, v)
	}
	slices.Sort(res)
	if !slices.Equal(res, []int{4, 9}) {
		t.Errorf("expected %v, got %v", []int{4, 9}, res)
	}
}

func TestPipeline_DoneBroadcast(t *testing.T) {
	done := make(chan struct{})

	// 输入放大到看门狗窗口内绝对跑不完的量级：
	// 这样"out 在 1 秒内关闭"只可能是 done 取消的功劳，不可能是自然跑完
	const total = 1_000_000
	nums := make([]int, total)
	for i := range nums {
		nums[i] = i
	}
	out := merge(done, sq(done, gen(done, nums...)))

	// 收到第一个结果，证明管线已转动（收不到就是组装错了，直接失败）
	select {
	case <-out:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output")
	}
	count := 1

	// 广播取消
	close(done)

	// 看门狗内收干 out：out 关闭 = 全管线退出完毕；
	// 取消后可能还有残余在途值（select 同时就绪时随机选），收下计数即可
	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-out:
			if !ok {
				// 手段二：收全了说明是自然跑完的，本测试什么都没证明，判失败
				if count == total {
					t.Fatal("received all values: pipeline finished naturally, cancel path not exercised")
				}
				t.Logf("cancelled after %d/%d values", count, total)
				return
			}
			count++
		case <-deadline:
			t.Fatal("timeout waiting for channel to close after done")
		}
	}
}

func TestWorkerPool(t *testing.T) {
	done := make(chan struct{})
	defer close(done)

	jobs := make(chan int)
	go func() {
		defer close(jobs)
		for i := 1; i <= 100; i++ {
			jobs <- i
		}
	}()
	out := workerPool(done, jobs, 4)
	res := make([]int, 0, 100)
	for v := range out {
		res = append(res, v)
	}
	want := make([]int, 100)
	for i := 1; i <= 100; i++ {
		want[i-1] = i * i
	}
	slices.Sort(res)
	if !slices.Equal(res, want) {
		t.Errorf("expected %v, got %v", want, res)
	}
}

func TestRateLimit(t *testing.T) {
	ticker := time.NewTicker(time.Millisecond * 20)
	defer ticker.Stop()

	stop := make(chan struct{})
	defer close(stop)

	jobs := make(chan int)
	go func() {
		i := 0
	loop1:
		for {
			select {
			case <-stop:
				break loop1
			case jobs <- i:
				i++
			}
		}
	}()

	count := 0
	deadline := time.After(time.Millisecond * 100)
loop2:
	for {
		select { //防抖，下方select在deadline和jobs都就绪的时候会随机选择执行，不能第一时间结束
		case <-deadline:
			break loop2
		default:
		}

		select {
		case <-deadline:
			break loop2
		case <-jobs:
			<-ticker.C
			count++
		}
	}
	if !(count > 0 && count <= 6) {
		t.Errorf("unexpected, got %d", count)
	}
}
