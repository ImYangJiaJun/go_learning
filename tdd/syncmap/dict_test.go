package syncmap

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestDictStoreLoad(t *testing.T) {
	var d Dict
	d.Store("lang", "go")

	v, ok := d.Load("lang")
	if !ok {
		t.Error("Expected to find key 'lang'")
	}
	if v != "go" {
		t.Errorf("Expected value 'go', got %v", v)
	}

	v, ok = d.Load("missing")
	if ok {
		t.Error("Expected ok to be false for missing key")
	}
	if v != nil {
		t.Errorf("Expected nil value for missing key, got %v", v)
	}
}

func TestLoadOrStore(t *testing.T) {
	t.Run("键不存在则存入", func(t *testing.T) {
		var d Dict
		actual, loaded := d.LoadOrStore("a", 1)
		if loaded {
			t.Error("Expected loaded to be false for new key")
		}
		if actual != 1 {
			t.Errorf("Expected actual 1, got %v", actual)
		}

		v, ok := d.Load("a")
		if !ok {
			t.Error("Expected to find key 'a' after LoadOrStore")
		}
		if v != 1 {
			t.Errorf("Expected value 1, got %v", v)
		}
	})

	t.Run("键已存在返回旧值", func(t *testing.T) {
		var d Dict
		d.Store("a", 1)
		actual, loaded := d.LoadOrStore("a", 999)
		if !loaded {
			t.Error("Expected loaded to be true for existing key")
		}
		if actual != 1 {
			t.Errorf("Expected old value 1, got %v", actual)
		}

		v, ok := d.Load("a")
		if !ok {
			t.Error("Expected to find key 'a'")
		}
		if v != 1 {
			t.Errorf("Expected value still 1 (999 discarded), got %v", v)
		}
	})
}

func TestDictDeleteRange(t *testing.T) {
	t.Run("Delete 后 Load 失败", func(t *testing.T) {
		var d Dict
		d.Store("a", 1)
		d.Delete("a")

		v, ok := d.Load("a")
		if ok {
			t.Error("Expected ok to be false after Delete")
		}
		if v != nil {
			t.Errorf("Expected nil value after Delete, got %v", v)
		}
	})

	t.Run("Delete 不存在的键", func(t *testing.T) {
		var d Dict
		d.Delete("ghost") // 不应 panic
		if got := d.Len(); got != 0 {
			t.Errorf("Expected Len 0, got %d", got)
		}
	})

	t.Run("Range 收集全部", func(t *testing.T) {
		var d Dict
		d.Store("a", 1)
		d.Store("b", 2)
		d.Store("c", 3)

		got := make(map[string]int)
		d.Range(func(key, value any) bool {
			k, ok := key.(string)
			if !ok {
				t.Fatalf("Expected string key, got %T", key)
			}
			v, ok := value.(int)
			if !ok {
				t.Fatalf("Expected int value, got %T", value)
			}
			got[k] = v
			return true
		})

		want := map[string]int{"a": 1, "b": 2, "c": 3}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %v, got %v", want, got)
		}
	})

	t.Run("Range 提前终止", func(t *testing.T) {
		var d Dict
		d.Store("a", 1)
		d.Store("b", 2)
		d.Store("c", 3)

		calls := 0
		got := make(map[string]int)
		d.Range(func(key, value any) bool {
			calls++
			k, ok := key.(string)
			if !ok {
				t.Fatalf("Expected string key, got %T", key)
			}
			v, ok := value.(int)
			if !ok {
				t.Fatalf("Expected int value, got %T", value)
			}
			got[k] = v
			return false
		})

		if calls != 1 {
			t.Errorf("Expected callback called exactly once, got %d", calls)
		}
		if len(got) >= 3 {
			t.Errorf("Expected fewer than 3 items collected, got %d", len(got))
		}
	})
}

func TestDictCompareAndSwap(t *testing.T) {
	t.Run("旧值匹配则替换", func(t *testing.T) {
		var d Dict
		d.Store("n", 1)
		if !d.CompareAndSwap("n", 1, 2) {
			t.Error("Expected CompareAndSwap to return true")
		}

		v, ok := d.Load("n")
		if !ok {
			t.Error("Expected to find key 'n'")
		}
		if v != 2 {
			t.Errorf("Expected value 2, got %v", v)
		}
	})

	t.Run("旧值不匹配不动", func(t *testing.T) {
		var d Dict
		d.Store("n", 1)
		if d.CompareAndSwap("n", 999, 2) {
			t.Error("Expected CompareAndSwap to return false")
		}

		v, ok := d.Load("n")
		if !ok {
			t.Error("Expected to find key 'n'")
		}
		if v != 1 {
			t.Errorf("Expected value still 1, got %v", v)
		}
	})

	t.Run("键不存在", func(t *testing.T) {
		var d Dict
		if d.CompareAndSwap("n", 1, 2) {
			t.Error("Expected CompareAndSwap to return false")
		}

		v, ok := d.Load("n")
		if ok {
			t.Error("Expected ok to be false for missing key")
		}
		if v != nil {
			t.Errorf("Expected nil value for missing key, got %v", v)
		}
	})
}

func TestDictConcurrentStore(t *testing.T) {
	var d Dict
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			for j := range 100 {
				d.Store(fmt.Sprintf("key-%d/%d", i, j), fmt.Sprintf("value-%d/%d", i, j))
			}
		})
	}
	wg.Wait()
	count := d.Len()
	if count != 1000 {
		t.Errorf("Expected 1000 items, got %d", count)
	}
}
