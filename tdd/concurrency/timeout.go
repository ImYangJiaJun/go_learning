package concurrency

import "time"

// FetchWithTimeout 在时限 d 内等待 ch 上的值：
// 收到则返回 (值, true)；超时返回 ("", false)，d 从调用时刻起算。
// ch 是只读定向 channel——接收方无权发送，更无权关闭它（关闭权属于发送方）。
func FetchWithTimeout(ch <-chan string, d time.Duration) (string, bool) {
	select {
	case res := <-ch:
		return res, true
	case <-time.After(d):
		return "", false
	}
}
