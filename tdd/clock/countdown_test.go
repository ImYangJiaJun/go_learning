package clock

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func TestCountdown(t *testing.T) {
	buffer := &bytes.Buffer{}
	Countdown(buffer, func(t time.Duration) {
		_, err := fmt.Fprintf(buffer, "sleep %v\n", t)
		if err != nil {
			return
		}
	})
	want := `3
sleep 1s
2
sleep 1s
1
sleep 1s
Go!
`
	if buffer.String() != want {
		t.Errorf("Countdown()=%v, want %v", buffer, want)
	}
}

// SyncBuffer 给 bytes.Buffer 加锁，供跨 goroutine 读写的测试使用。
type SyncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *SyncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *SyncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

/*
synctest
气泡：synctest.Test 圈出隔离世界，泡内 goroutine 共享假时钟（起点 2000-01-01）
时间前进条件：泡内所有 goroutine 都持久阻塞（泡内 channel/Sleep/WaitGroup；Mutex、I/O 不算）
测试里的 time.Sleep 不真睡 = 把假钟往前拧
synctest.Wait()：等泡内其他 goroutine 全部睡死，断言前的同步点，防抖动
Test 返回前等泡内 goroutine 全退场，泄漏则死锁 panic（白捡泄漏检测）
*/
func TestCountdownWithRealSleeper(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// 跨 goroutine 读写，必须加锁（裸 bytes.Buffer 会被 -race 抓）
		buffer := &SyncBuffer{}

		// 传真的 time.Sleep——泡内睡的是假钟，不花真实时间
		go Countdown(buffer, time.Sleep)

		// 等倒计时 goroutine 写完 "3" 并睡死，再断言
		synctest.Wait()
		assertOutput(t, buffer, "3\n")

		// 假钟 +1s：倒计时醒来写 "2"，再睡死
		time.Sleep(1 * time.Second)
		synctest.Wait()
		assertOutput(t, buffer, "3\n2\n")

		time.Sleep(1 * time.Second)
		synctest.Wait()
		assertOutput(t, buffer, "3\n2\n1\n")

		// 假钟 +3s：写 "Go!"，goroutine 退场
		time.Sleep(1 * time.Second)
		synctest.Wait()
		assertOutput(t, buffer, "3\n2\n1\nGo!\n")
	})
}

func assertOutput(t *testing.T, buffer *SyncBuffer, want string) {
	t.Helper()
	if got := buffer.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
