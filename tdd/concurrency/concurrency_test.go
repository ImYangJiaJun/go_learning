package concurrency

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClose(t *testing.T) {
	t.Run("已关闭后接收", func(t *testing.T) {
		var c = make(chan string, 3)
		close(c)
		v, ok := <-c
		if v != "" || ok {
			t.Errorf("接收已关闭的通道应该返回零值和false, got %v, %v", v, ok)
		}
	})

	t.Run("重复 close", func(t *testing.T) {
		var c = make(chan string, 3)
		close(c)
		defer func() {
			if r := recover(); !strings.Contains(fmt.Sprint(r), "close of closed channel") {
				t.Errorf("重复 close 应该 panic, got %v", r)
			}
		}()
		close(c)
	})

	t.Run("向已关闭发送", func(t *testing.T) {
		var c = make(chan string, 3)
		close(c)
		defer func() {
			if r := recover(); !strings.Contains(fmt.Sprint(r), "send on closed channel") {
				t.Errorf("向已关闭的通道发送应该 panic, got %v", r)
			}
		}()
		c <- "test"
	})
}

func TestUnbufferedChannelBlocksUntilReceived(t *testing.T) {
	ch := make(chan string)     // 注意：无缓冲
	sent := make(chan struct{}) // 事件信号：发送完成才 close

	go func() {
		ch <- "x"   // 这一行会阻塞，直到有人接收
		close(sent) // 能执行到这里 = 发送已完成 = 阻塞已解除
	}()

	// 第一幕：故意不接收 → 断言"该阻塞"
	// select 超时惯用法：事件 case 与 time.After 闹钟竞争，
	// 闹钟先响 = 期限内事件没发生（负断言，故用短超时）
	select {
	case <-sent:
		t.Error("没人接收，发送方不该完成")
	case <-time.After(50 * time.Millisecond): // time.After 返回 d 后才送值的 channel
		// 期望走到这里：50ms 内 sent 未被 close，证明发送阻塞了
	}

	// 第二幕：接收 → 断言"该解除"
	got := <-ch
	if got != "x" {
		t.Errorf("got %q, want %q", got, "x")
	}
	select {
	case <-sent:
		// 期望走到这里：接收后发送方立即解除阻塞
	case <-time.After(1 * time.Second):
		t.Error("接收后发送方应该立刻完成")
	}
}

// slowCh 测试辅助：返回一个 channel，内部 goroutine 睡 delay 后才发送 v
// 注意：超时用例里这个 goroutine 会永远阻塞在发送上（无人接收）——超时就有泄漏代价
func slowCh(delay time.Duration, v string) <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(delay)
		ch <- v
	}()
	return ch
}

func TestFetchWithTimeout(t *testing.T) {
	// 表驱动：两行用例只有"谁先就绪"不同——delay 与 d 的相对大小决定 select 谁赢
	cases := []struct {
		name   string
		delay  time.Duration // slowCh 的发送延迟
		d      time.Duration // FetchWithTimeout 的超时时限
		want   string
		wantOK bool
	}{
		{"超时返回零值和 false", 50 * time.Millisecond, 10 * time.Millisecond, "", false}, // 闹钟先响
		{"时限内取到值", 50 * time.Millisecond, 1 * time.Second, "late", true},           // 值先到
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := FetchWithTimeout(slowCh(c.delay, "late"), c.d)
			if got != c.want || ok != c.wantOK {
				t.Errorf("FetchWithTimeout = (%q, %v), want (%q, %v)", got, ok, c.want, c.wantOK)
			}
		})
	}
}

func TestWaitGroup(t *testing.T) {
	t.Run("计数为负 panic", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		wg.Done()
		defer func() {
			if r := recover(); !strings.Contains(fmt.Sprint(r), "sync: negative WaitGroup counter") {
				t.Errorf("Done 多于 Add 应该 panic（计数为负）, got %v", r)
			}
		}()
		wg.Done() // 第二次 Done 使计数变负 → panic
	})

	t.Run("wg.Go 等价手写 Add/Done", func(t *testing.T) {
		done := make(chan int, 3) // 缓冲 channel 收集完成事件：缓冲有空位，任务不会因发送阻塞
		var wg sync.WaitGroup
		for i := range 3 {
			wg.Go(func() { done <- i }) // wg.Go = Add(1) + go + defer Done 一体（Go 1.25）
		}
		wg.Wait() // Wait 建立 happens-before：返回后读 done 的结果无需加锁
		if len(done) != 3 {
			t.Errorf("Wait 返回后 3 个任务的结果应全部可见, got len(done) = %d", len(done))
		}
	})
}
