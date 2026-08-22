package contextlab

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"testing/synctest"
	"time"
)

type slowStore struct {
	delay time.Duration
}

func (s slowStore) Get(id int) (string, error) {
	time.Sleep(s.delay)
	return fmt.Sprintf("item-%d", id), nil
}

type requestIDKey struct{}

func TestFetch(t *testing.T) {
	t.Run("存储慢于取消", func(t *testing.T) {
		ss := slowStore{delay: 200 * time.Millisecond}
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 创建即取消

		got, err := Fetch(ctx, ss, 1)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("父取消整棵子树", func(t *testing.T) {
		parent, cancelParent := context.WithCancel(context.Background())
		defer cancelParent()
		child, cancelChild := context.WithCancel(parent)
		defer cancelChild()
		grandchild, cancelGrandchild := context.WithCancel(child)
		defer cancelGrandchild()

		// cancel 前：整棵树都活着
		assertNotCanceled(t, "parent", parent)
		assertNotCanceled(t, "child", child)
		assertNotCanceled(t, "grandchild", grandchild)
		t.Log("cancel 前 parent.Err():", parent.Err())

		cancelParent() // 唯一的动作：只 cancel 父节点

		t.Log("cancel 后 parent.Err():", parent.Err())
		// cancel 是同步的：返回时级联已完成，不需要 sleep
		assertCanceled(t, "parent", parent)
		assertCanceled(t, "child", child)
		assertCanceled(t, "grandchild", grandchild)
	})

	t.Run("存储慢于超时", func(t *testing.T) {
		ss := slowStore{delay: 200 * time.Millisecond}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		got, err := Fetch(ctx, ss, 1)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("存储快于超时", func(t *testing.T) {
		ss := slowStore{delay: 50 * time.Millisecond}
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		got, err := Fetch(ctx, ss, 42)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "item-42" {
			t.Fatalf("expected 'item-42', got %q", got)
		}
	})

	t.Run("截止时间已过", func(t *testing.T) {
		ss := slowStore{delay: 10 * time.Millisecond}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second)) // 截止时间已过
		defer cancel()

		got, err := Fetch(ctx, ss, 1)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("截止时间将到", func(t *testing.T) {
		ss := slowStore{delay: 200 * time.Millisecond}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond)) // 截止时间将到
		defer cancel()

		got, err := Fetch(ctx, ss, 1)
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", err)
		}
	})

	t.Run("对照：未到期且存储快", func(t *testing.T) {
		ss := slowStore{delay: 50 * time.Millisecond}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(200*time.Millisecond)) // 截止时间未到
		defer cancel()

		got, err := Fetch(ctx, ss, 42)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "item-42" {
			t.Fatalf("expected 'item-42', got %q", got)
		}
	})

	t.Run("存后能取", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), requestIDKey{}, "req-123")
		got := ctx.Value(requestIDKey{})
		if got != "req-123" {
			t.Fatalf("expected 'req-123', got %v", got)
		}
	})

	t.Run("没存过", func(t *testing.T) {
		ctx := context.Background()
		got := ctx.Value(requestIDKey{})
		if got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("子覆盖父", func(t *testing.T) {
		parent := context.WithValue(context.Background(), requestIDKey{}, "req-123")
		child := context.WithValue(parent, requestIDKey{}, "req-456")
		gotParent := parent.Value(requestIDKey{})
		gotChild := child.Value(requestIDKey{})
		if gotParent != "req-123" {
			t.Fatalf("expected 'req-123' from parent, got %v", gotParent)
		}
		if gotChild != "req-456" {
			t.Fatalf("expected 'req-456' from child, got %v", gotChild)
		}
	})
}

// synctest 气泡：假时钟直接快进，200ms 存储、50ms 超时真实耗时≈0
func TestFetch_SynctestTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		got, err := Fetch(ctx, slowStore{delay: 200 * time.Millisecond}, 1)

		// 快进到存储醒来，让后台 worker 收尾——气泡不允许测试返回时还有 goroutine 活着
		time.Sleep(time.Second)

		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected context.DeadlineExceeded, got %v", err)
		}
	})
}

// assertCanceled 断言 ctx 此刻已被取消：Done() 已关闭且 Err() 为 context.Canceled。
func assertCanceled(t *testing.T, name string, ctx context.Context) {
	t.Helper()
	select {
	case <-ctx.Done():
	default:
		t.Errorf("%s 的 Done() 应该已关闭", name)
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Errorf("%s 的 Err() 应该是 context.Canceled，得到 %v", name, ctx.Err())
	}
}

// assertNotCanceled 断言 ctx 此刻尚未取消（Done() 仍开着，Err() 为 nil）。
func assertNotCanceled(t *testing.T, name string, ctx context.Context) {
	t.Helper()
	select {
	case <-ctx.Done():
		t.Errorf("%s 的 Done() 不应该关闭", name)
	default:
	}
	if err := ctx.Err(); err != nil {
		t.Errorf("%s 的 Err() 应该是 nil，得到 %v", name, err)
	}
}
