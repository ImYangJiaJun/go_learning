// 包名叫 context 会遮蔽标准库，所以叫 contextlab
package contextlab

import "context"

// Store 是慢存储的抽象：Get 可能耗时几百毫秒才返回。
// 实现方不负责响应取消——那是 Fetch 的事。
type Store interface {
	Get(id int) (string, error)
}

// Fetch 从慢存储查询 id 对应的数据。
// 必须尊重 ctx：内部用 select 同时等"存储结果"和"ctx.Done()"——
// 存储先返回，则返回存储结果；ctx 先 Done（取消/超时），则立即放弃等待，
// 返回 "", ctx.Err()（context.Canceled 或 context.DeadlineExceeded）。
func Fetch(ctx context.Context, s Store, id int) (string, error) {
	type result struct {
		res string
		err error
	}
	ch := make(chan result, 1)
	go func() {
		v, err := s.Get(id)
		ch <- result{v, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-ch:
		return res.res, res.err
	}
}
